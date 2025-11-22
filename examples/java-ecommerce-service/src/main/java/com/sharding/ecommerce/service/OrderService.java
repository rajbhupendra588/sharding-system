package com.sharding.ecommerce.service;

import com.sharding.ecommerce.model.Order;
import com.sharding.ecommerce.model.OrderItem;
import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.ShardingClientException;
import com.sharding.system.client.model.QueryResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.stream.Collectors;

/**
 * Order Service demonstrating sharding by user ID.
 * 
 * Orders are sharded by user_id, ensuring:
 * - All orders for a user are on the same shard (co-location)
 * - Queries like "get all orders for user X" are fast (single shard)
 * - Order creation and user updates can be transactional within a shard
 * - Efficient order history queries without cross-shard operations
 */
@Service
public class OrderService {

    private static final Logger log = LoggerFactory.getLogger(OrderService.class);
    private final ShardingClient shardingClient;

    public OrderService(ShardingClient shardingClient) {
        this.shardingClient = shardingClient;
    }

    /**
     * Creates a new order.
     * Uses user_id as the shard key to ensure orders are co-located with user data.
     */
    public Order createOrder(Order order) throws ShardingClientException {
        log.info("Creating order for user: {}", order.getUserId());
        
        String orderId = order.getId() != null ? order.getId() : UUID.randomUUID().toString();
        order.setId(orderId);
        order.setCreatedAt(LocalDateTime.now());
        order.setUpdatedAt(LocalDateTime.now());
        
        String userId = order.getUserId(); // Shard key
        
        // Insert order header
        shardingClient.queryStrong(
            userId, // Shard key - ensures order is on same shard as user
            """
                INSERT INTO orders (id, user_id, total_amount, status, shipping_address, 
                                  payment_method, tracking_number, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            """,
            orderId,
            userId,
            order.getTotalAmount(),
            order.getStatus().name(),
            order.getShippingAddress(),
            order.getPaymentMethod(),
            order.getTrackingNumber(),
            order.getCreatedAt(),
            order.getUpdatedAt()
        );
        
        // Insert order items
        if (order.getItems() != null && !order.getItems().isEmpty()) {
            for (OrderItem item : order.getItems()) {
                shardingClient.queryStrong(
                    userId, // Same shard key
                    """
                        INSERT INTO order_items (order_id, product_id, product_name, quantity, 
                                               unit_price, total_price)
                        VALUES ($1, $2, $3, $4, $5, $6)
                    """,
                    orderId,
                    item.getProductId(),
                    item.getProductName(),
                    item.getQuantity(),
                    item.getUnitPrice(),
                    item.getTotalPrice()
                );
            }
        }
        
        log.info("Order created successfully. Order ID: {}, Shard key: {}", orderId, userId);
        return order;
    }

    /**
     * Retrieves an order by ID.
     * Uses user_id from the order to route to the correct shard.
     */
    public Order getOrderById(String orderId, String userId) throws ShardingClientException {
        log.debug("Fetching order {} for user {}", orderId, userId);
        
        QueryResponse response = shardingClient.queryStrong(
            userId, // Shard key
            """
                SELECT o.id, o.user_id, o.total_amount, o.status, o.shipping_address,
                       o.payment_method, o.tracking_number, o.created_at, o.updated_at,
                       o.shipped_at, o.delivered_at
                FROM orders o
                WHERE o.id = $1 AND o.user_id = $2
            """,
            orderId,
            userId
        );
        
        if (response.getRowCount() == 0) {
            log.warn("Order not found: {} for user {}", orderId, userId);
            return null;
        }
        
        Order order = mapRowToOrder(response.getRows().get(0));
        
        // Fetch order items
        order.setItems(getOrderItems(orderId, userId));
        
        return order;
    }

    /**
     * Gets all orders for a user.
     * This is efficient because all orders are on the same shard as the user.
     */
    public List<Order> getOrdersByUserId(String userId) throws ShardingClientException {
        log.debug("Fetching all orders for user: {}", userId);
        
        QueryResponse response = shardingClient.queryStrong(
            userId, // Shard key
            """
                SELECT o.id, o.user_id, o.total_amount, o.status, o.shipping_address,
                       o.payment_method, o.tracking_number, o.created_at, o.updated_at,
                       o.shipped_at, o.delivered_at
                FROM orders o
                WHERE o.user_id = $1
                ORDER BY o.created_at DESC
            """,
            userId
        );
        
        List<Order> orders = response.getRows().stream()
            .map(this::mapRowToOrder)
            .collect(Collectors.toList());
        
        // Fetch items for each order
        for (Order order : orders) {
            order.setItems(getOrderItems(order.getId(), userId));
        }
        
        log.debug("Found {} orders for user {}", orders.size(), userId);
        return orders;
    }

    /**
     * Updates order status.
     */
    public Order updateOrderStatus(String orderId, String userId, Order.OrderStatus status) 
            throws ShardingClientException {
        log.info("Updating order {} status to {} for user {}", orderId, status, userId);
        
        LocalDateTime now = LocalDateTime.now();
        LocalDateTime shippedAt = status == Order.OrderStatus.SHIPPED ? now : null;
        LocalDateTime deliveredAt = status == Order.OrderStatus.DELIVERED ? now : null;
        
        shardingClient.queryStrong(
            userId, // Shard key
            """
                UPDATE orders 
                SET status = $3, updated_at = $4, shipped_at = $5, delivered_at = $6
                WHERE id = $1 AND user_id = $2
            """,
            orderId,
            userId,
            status.name(),
            now,
            shippedAt,
            deliveredAt
        );
        
        return getOrderById(orderId, userId);
    }

    /**
     * Gets order statistics for a user.
     * Efficient because all data is on the same shard.
     */
    public Map<String, Object> getOrderStatistics(String userId) throws ShardingClientException {
        log.debug("Getting order statistics for user: {}", userId);
        
        QueryResponse response = shardingClient.queryStrong(
            userId, // Shard key
            """
                SELECT 
                    COUNT(*) as total_orders,
                    SUM(total_amount) as total_spent,
                    AVG(total_amount) as avg_order_value,
                    MAX(created_at) as last_order_date
                FROM orders
                WHERE user_id = $1
            """,
            userId
        );
        
        if (response.getRowCount() == 0) {
            return Map.of(
                "total_orders", 0,
                "total_spent", BigDecimal.ZERO,
                "avg_order_value", BigDecimal.ZERO,
                "last_order_date", null
            );
        }
        
        return response.getRows().get(0);
    }

    private List<OrderItem> getOrderItems(String orderId, String userId) throws ShardingClientException {
        QueryResponse response = shardingClient.queryStrong(
            userId, // Same shard key
            """
                SELECT product_id, product_name, quantity, unit_price, total_price
                FROM order_items
                WHERE order_id = $1
            """,
            orderId
        );
        
        return response.getRows().stream()
            .map(row -> OrderItem.builder()
                .productId((String) row.get("product_id"))
                .productName((String) row.get("product_name"))
                .quantity((Integer) row.get("quantity"))
                .unitPrice((BigDecimal) row.get("unit_price"))
                .totalPrice((BigDecimal) row.get("total_price"))
                .build())
            .collect(Collectors.toList());
    }

    private Order mapRowToOrder(Map<String, Object> row) {
        return Order.builder()
            .id((String) row.get("id"))
            .userId((String) row.get("user_id"))
            .totalAmount((BigDecimal) row.get("total_amount"))
            .status(Order.OrderStatus.valueOf((String) row.get("status")))
            .shippingAddress((String) row.get("shipping_address"))
            .paymentMethod((String) row.get("payment_method"))
            .trackingNumber((String) row.get("tracking_number"))
            .createdAt((LocalDateTime) row.get("created_at"))
            .updatedAt((LocalDateTime) row.get("updated_at"))
            .shippedAt((LocalDateTime) row.get("shipped_at"))
            .deliveredAt((LocalDateTime) row.get("delivered_at"))
            .build();
    }
}

