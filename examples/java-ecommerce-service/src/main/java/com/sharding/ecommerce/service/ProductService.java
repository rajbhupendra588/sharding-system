package com.sharding.ecommerce.service;

import com.sharding.ecommerce.model.Product;
import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.ShardingClientException;
import com.sharding.system.client.model.QueryResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.stream.Collectors;

/**
 * Product Service demonstrating sharding by product ID.
 * 
 * Products are sharded by product_id. For production, you might want to:
 * - Use a separate catalog shard for products
 * - Replicate product data across shards for read performance
 * - Use eventual consistency for product reads (they change less frequently)
 */
@Service
public class ProductService {

    private static final Logger log = LoggerFactory.getLogger(ProductService.class);
    private final ShardingClient shardingClient;

    public ProductService(ShardingClient shardingClient) {
        this.shardingClient = shardingClient;
    }

    /**
     * Creates a new product.
     * Uses product_id as the shard key.
     */
    public Product createProduct(Product product) throws ShardingClientException {
        log.info("Creating product with ID: {}", product.getId());
        
        String productId = product.getId() != null ? product.getId() : UUID.randomUUID().toString();
        product.setId(productId);
        product.setCreatedAt(LocalDateTime.now());
        product.setUpdatedAt(LocalDateTime.now());
        
        shardingClient.queryStrong(
            productId, // Shard key
            """
                INSERT INTO products (id, name, description, price, stock_quantity, 
                                   category, brand, image_url, created_at, updated_at, active)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
            """,
            productId,
            product.getName(),
            product.getDescription(),
            product.getPrice(),
            product.getStockQuantity(),
            product.getCategory(),
            product.getBrand(),
            product.getImageUrl(),
            product.getCreatedAt(),
            product.getUpdatedAt(),
            product.getActive() != null ? product.getActive() : true
        );
        
        log.info("Product created successfully. Shard key: {}", productId);
        return product;
    }

    /**
     * Retrieves a product by ID.
     * Uses eventual consistency for better read performance (products change less frequently).
     */
    public Product getProductById(String productId) throws ShardingClientException {
        log.debug("Fetching product with ID: {}", productId);
        
        QueryResponse response = shardingClient.queryEventual(
            productId, // Shard key - eventual consistency for read performance
            """
                SELECT id, name, description, price, stock_quantity, category, 
                       brand, image_url, created_at, updated_at, active
                FROM products
                WHERE id = $1 AND active = true
            """,
            productId
        );
        
        if (response.getRowCount() == 0) {
            log.warn("Product not found: {}", productId);
            return null;
        }
        
        return mapRowToProduct(response.getRows().get(0));
    }

    /**
     * Updates product information.
     */
    public Product updateProduct(String productId, Product product) throws ShardingClientException {
        log.info("Updating product with ID: {}", productId);
        
        product.setUpdatedAt(LocalDateTime.now());
        
        shardingClient.queryStrong(
            productId, // Shard key
            """
                UPDATE products 
                SET name = $2, description = $3, price = $4, stock_quantity = $5,
                    category = $6, brand = $7, image_url = $8, updated_at = $9, active = $10
                WHERE id = $1
            """,
            productId,
            product.getName(),
            product.getDescription(),
            product.getPrice(),
            product.getStockQuantity(),
            product.getCategory(),
            product.getBrand(),
            product.getImageUrl(),
            product.getUpdatedAt(),
            product.getActive() != null ? product.getActive() : true
        );
        
        return getProductById(productId);
    }

    /**
     * Updates product stock quantity.
     * Important: In production, use optimistic locking or transactions to prevent race conditions.
     */
    public void updateStock(String productId, int quantityChange) throws ShardingClientException {
        log.info("Updating stock for product {} by {}", productId, quantityChange);
        
        shardingClient.queryStrong(
            productId, // Shard key
            """
                UPDATE products 
                SET stock_quantity = stock_quantity + $2, updated_at = $3
                WHERE id = $1
            """,
            productId,
            quantityChange,
            LocalDateTime.now()
        );
    }

    /**
     * Gets products by category.
     * Note: This requires scanning multiple shards. For production, consider:
     * - Using a search index (Elasticsearch)
     * - Maintaining a separate catalog shard
     * - Using eventual consistency for better performance
     */
    public List<Product> getProductsByCategory(String category) throws ShardingClientException {
        log.debug("Fetching products by category: {}", category);
        
        // Note: This is a simplified example. In production, you'd want to:
        // 1. Query all shards and aggregate results
        // 2. Use a search index
        // 3. Maintain a catalog shard
        
        // For demo purposes, we'll query using a generic key
        // In real production, implement multi-shard query aggregation
        QueryResponse response = shardingClient.queryEventual(
            category, // Using category as shard key for demo (not ideal)
            """
                SELECT id, name, description, price, stock_quantity, category, 
                       brand, image_url, created_at, updated_at, active
                FROM products
                WHERE category = $1 AND active = true
                LIMIT 100
            """,
            category
        );
        
        return response.getRows().stream()
            .map(this::mapRowToProduct)
            .collect(Collectors.toList());
    }

    private Product mapRowToProduct(Map<String, Object> row) {
        return Product.builder()
            .id((String) row.get("id"))
            .name((String) row.get("name"))
            .description((String) row.get("description"))
            .price((java.math.BigDecimal) row.get("price"))
            .stockQuantity((Integer) row.get("stock_quantity"))
            .category((String) row.get("category"))
            .brand((String) row.get("brand"))
            .imageUrl((String) row.get("image_url"))
            .createdAt((LocalDateTime) row.get("created_at"))
            .updatedAt((LocalDateTime) row.get("updated_at"))
            .active((Boolean) row.get("active"))
            .build();
    }
}

