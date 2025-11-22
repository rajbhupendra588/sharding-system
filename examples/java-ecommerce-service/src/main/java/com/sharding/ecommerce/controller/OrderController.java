package com.sharding.ecommerce.controller;

import com.sharding.ecommerce.model.Order;
import com.sharding.ecommerce.service.OrderService;
import com.sharding.system.client.ShardingClientException;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

/**
 * REST Controller for Order operations.
 * 
 * Demonstrates how sharding by user_id enables:
 * - Fast order history queries (all orders on same shard as user)
 * - Efficient order creation (co-located with user data)
 * - No cross-shard queries for user order operations
 */
@RestController
@RequestMapping("/api/v1/orders")
@Tag(name = "Orders", description = "Order management API")
public class OrderController {

    private static final Logger log = LoggerFactory.getLogger(OrderController.class);
    private final OrderService orderService;

    public OrderController(OrderService orderService) {
        this.orderService = orderService;
    }

    @PostMapping
    @Operation(summary = "Create a new order", 
               description = "Creates a new order. Uses user_id as shard key to co-locate with user data.")
    public ResponseEntity<Order> createOrder(@Valid @RequestBody Order order) {
        try {
            Order created = orderService.createOrder(order);
            return ResponseEntity.status(HttpStatus.CREATED).body(created);
        } catch (ShardingClientException e) {
            log.error("Error creating order", e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @GetMapping("/{orderId}")
    @Operation(summary = "Get order by ID", 
               description = "Retrieves an order. Requires user_id to route to correct shard.")
    public ResponseEntity<Order> getOrderById(
            @Parameter(description = "Order ID", required = true) @PathVariable String orderId,
            @Parameter(description = "User ID (shard key)", required = true) @RequestParam String userId) {
        try {
            Order order = orderService.getOrderById(orderId, userId);
            if (order == null) {
                return ResponseEntity.notFound().build();
            }
            return ResponseEntity.ok(order);
        } catch (ShardingClientException e) {
            log.error("Error fetching order: {}", orderId, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @GetMapping("/user/{userId}")
    @Operation(summary = "Get all orders for a user", 
               description = "Retrieves all orders for a user. Efficient because all orders are on the same shard.")
    public ResponseEntity<List<Order>> getOrdersByUserId(
            @Parameter(description = "User ID (shard key)", required = true) @PathVariable String userId) {
        try {
            List<Order> orders = orderService.getOrdersByUserId(userId);
            return ResponseEntity.ok(orders);
        } catch (ShardingClientException e) {
            log.error("Error fetching orders for user: {}", userId, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @PutMapping("/{orderId}/status")
    @Operation(summary = "Update order status", 
               description = "Updates order status (e.g., SHIPPED, DELIVERED). Uses user_id as shard key.")
    public ResponseEntity<Order> updateOrderStatus(
            @Parameter(description = "Order ID", required = true) @PathVariable String orderId,
            @Parameter(description = "User ID (shard key)", required = true) @RequestParam String userId,
            @RequestBody Map<String, String> statusUpdate) {
        try {
            Order.OrderStatus status = Order.OrderStatus.valueOf(statusUpdate.get("status"));
            Order updated = orderService.updateOrderStatus(orderId, userId, status);
            return ResponseEntity.ok(updated);
        } catch (IllegalArgumentException e) {
            log.error("Invalid order status", e);
            return ResponseEntity.badRequest().build();
        } catch (ShardingClientException e) {
            log.error("Error updating order status: {}", orderId, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @GetMapping("/user/{userId}/statistics")
    @Operation(summary = "Get order statistics for a user", 
               description = "Returns order statistics (total orders, total spent, etc.). Efficient single-shard query.")
    public ResponseEntity<Map<String, Object>> getOrderStatistics(
            @Parameter(description = "User ID (shard key)", required = true) @PathVariable String userId) {
        try {
            Map<String, Object> stats = orderService.getOrderStatistics(userId);
            return ResponseEntity.ok(stats);
        } catch (ShardingClientException e) {
            log.error("Error fetching order statistics for user: {}", userId, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }
}

