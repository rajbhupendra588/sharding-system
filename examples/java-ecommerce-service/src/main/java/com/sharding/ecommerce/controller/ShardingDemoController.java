package com.sharding.ecommerce.controller;

import com.sharding.ecommerce.service.UserService;
import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.ShardingClientException;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.*;

/**
 * Demo Controller showcasing sharding system capabilities.
 * 
 * This controller demonstrates:
 * - How data is distributed across shards
 * - Shard key routing
 * - Benefits of co-location
 * - Performance characteristics
 */
@RestController
@RequestMapping("/api/v1/demo")
@Tag(name = "Sharding Demo", description = "Demonstrates sharding system capabilities")
public class ShardingDemoController {

    private static final Logger log = LoggerFactory.getLogger(ShardingDemoController.class);
    private final ShardingClient shardingClient;
    private final UserService userService;

    public ShardingDemoController(ShardingClient shardingClient, UserService userService) {
        this.shardingClient = shardingClient;
        this.userService = userService;
    }

    @GetMapping("/shard-distribution")
    @Operation(summary = "Show shard distribution", 
               description = "Demonstrates how different keys map to different shards")
    public ResponseEntity<Map<String, Object>> showShardDistribution(
            @RequestParam(defaultValue = "10") int sampleSize) {
        try {
            Map<String, String> distribution = new LinkedHashMap<>();
            Map<String, Integer> shardCounts = new HashMap<>();
            
            // Generate sample keys and show their shard mapping
            for (int i = 0; i < sampleSize; i++) {
                String key = "user-" + UUID.randomUUID().toString();
                String shardId = shardingClient.getShardForKey(key);
                distribution.put(key, shardId);
                shardCounts.put(shardId, shardCounts.getOrDefault(shardId, 0) + 1);
            }
            
            Map<String, Object> result = new HashMap<>();
            result.put("sample_keys", distribution);
            result.put("shard_distribution", shardCounts);
            result.put("total_keys", sampleSize);
            result.put("unique_shards", shardCounts.size());
            result.put("message", "Keys are distributed across " + shardCounts.size() + " shards");
            
            return ResponseEntity.ok(result);
        } catch (ShardingClientException e) {
            log.error("Error showing shard distribution", e);
            return ResponseEntity.internalServerError().build();
        }
    }

    @GetMapping("/benefits")
    @Operation(summary = "Explain sharding benefits", 
               description = "Returns information about how sharding helps this application")
    public ResponseEntity<Map<String, Object>> explainBenefits() {
        Map<String, Object> benefits = new LinkedHashMap<>();
        
        benefits.put("horizontal_scaling", Map.of(
            "description", "Scale database capacity by adding more shards",
            "benefit", "Handle millions of users without single database bottleneck",
            "example", "Add shard-03, shard-04 as user base grows"
        ));
        
        benefits.put("performance", Map.of(
            "description", "Queries only hit one shard instead of scanning entire database",
            "benefit", "Faster query execution, lower latency",
            "example", "Get user orders: O(1) shard lookup vs O(n) full table scan"
        ));
        
        benefits.put("co_location", Map.of(
            "description", "Related data stored on same shard",
            "benefit", "Efficient joins and transactions within a shard",
            "example", "User and their orders on same shard - fast order history queries"
        ));
        
        benefits.put("fault_isolation", Map.of(
            "description", "Shard failures don't affect entire system",
            "benefit", "Better availability and resilience",
            "example", "If shard-01 fails, only users on that shard are affected"
        ));
        
        benefits.put("resharding", Map.of(
            "description", "Split or merge shards without downtime",
            "benefit", "Adapt to changing data distribution patterns",
            "example", "Split hot shard into multiple shards when it grows too large"
        ));
        
        Map<String, Object> response = new HashMap<>();
        response.put("benefits", benefits);
        response.put("sharding_strategy", "Consistent hashing with virtual nodes");
        response.put("shard_key_examples", Map.of(
            "users", "user_id",
            "orders", "user_id (co-located with users)",
            "products", "product_id"
        ));
        
        return ResponseEntity.ok(response);
    }

    @GetMapping("/shard-for-key/{key}")
    @Operation(summary = "Get shard for a key", 
               description = "Shows which shard contains data for a given key")
    public ResponseEntity<Map<String, String>> getShardForKey(@PathVariable String key) {
        try {
            String shardId = shardingClient.getShardForKey(key);
            return ResponseEntity.ok(Map.of(
                "key", key,
                "shard_id", shardId,
                "message", "Data for key '" + key + "' is stored on shard '" + shardId + "'"
            ));
        } catch (ShardingClientException e) {
            log.error("Error getting shard for key: {}", key, e);
            return ResponseEntity.internalServerError().build();
        }
    }
}

