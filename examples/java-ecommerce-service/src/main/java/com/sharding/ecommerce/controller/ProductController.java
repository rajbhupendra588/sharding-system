package com.sharding.ecommerce.controller;

import com.sharding.ecommerce.model.Product;
import com.sharding.ecommerce.service.ProductService;
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
 * REST Controller for Product operations.
 * 
 * Demonstrates product sharding by product_id.
 * Note: For production, consider using eventual consistency for reads
 * and maintaining a separate catalog shard or search index.
 */
@RestController
@RequestMapping("/api/v1/products")
@Tag(name = "Products", description = "Product catalog API")
public class ProductController {

    private static final Logger log = LoggerFactory.getLogger(ProductController.class);
    private final ProductService productService;

    public ProductController(ProductService productService) {
        this.productService = productService;
    }

    @PostMapping
    @Operation(summary = "Create a new product", description = "Creates a new product. Uses product_id as shard key.")
    public ResponseEntity<Product> createProduct(@Valid @RequestBody Product product) {
        try {
            Product created = productService.createProduct(product);
            return ResponseEntity.status(HttpStatus.CREATED).body(created);
        } catch (ShardingClientException e) {
            log.error("Error creating product", e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @GetMapping("/{id}")
    @Operation(summary = "Get product by ID", 
               description = "Retrieves product information. Uses eventual consistency for better read performance.")
    public ResponseEntity<Product> getProductById(
            @Parameter(description = "Product ID", required = true) @PathVariable String id) {
        try {
            Product product = productService.getProductById(id);
            if (product == null) {
                return ResponseEntity.notFound().build();
            }
            return ResponseEntity.ok(product);
        } catch (ShardingClientException e) {
            log.error("Error fetching product: {}", id, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @PutMapping("/{id}")
    @Operation(summary = "Update product", description = "Updates product information.")
    public ResponseEntity<Product> updateProduct(
            @Parameter(description = "Product ID", required = true) @PathVariable String id,
            @Valid @RequestBody Product product) {
        try {
            Product updated = productService.updateProduct(id, product);
            return ResponseEntity.ok(updated);
        } catch (ShardingClientException e) {
            log.error("Error updating product: {}", id, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @PutMapping("/{id}/stock")
    @Operation(summary = "Update product stock", 
               description = "Updates product stock quantity. Uses optimistic locking in production.")
    public ResponseEntity<Void> updateStock(
            @Parameter(description = "Product ID", required = true) @PathVariable String id,
            @RequestBody Map<String, Integer> stockUpdate) {
        try {
            Integer quantityChange = stockUpdate.get("quantity_change");
            if (quantityChange == null) {
                return ResponseEntity.badRequest().build();
            }
            productService.updateStock(id, quantityChange);
            return ResponseEntity.noContent().build();
        } catch (ShardingClientException e) {
            log.error("Error updating stock for product: {}", id, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @GetMapping("/category/{category}")
    @Operation(summary = "Get products by category", 
               description = "Retrieves products by category. Note: In production, use a search index for better performance.")
    public ResponseEntity<List<Product>> getProductsByCategory(
            @Parameter(description = "Product category", required = true) @PathVariable String category) {
        try {
            List<Product> products = productService.getProductsByCategory(category);
            return ResponseEntity.ok(products);
        } catch (ShardingClientException e) {
            log.error("Error fetching products by category: {}", category, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }
}

