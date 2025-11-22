package com.sharding.ecommerce.health;

import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.ShardingClientException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.HealthIndicator;
import org.springframework.stereotype.Component;

/**
 * Custom health indicator for sharding system connectivity.
 * 
 * Checks if the sharding router is accessible and responding.
 */
@Component
public class ShardingHealthIndicator implements HealthIndicator {

    private static final Logger log = LoggerFactory.getLogger(ShardingHealthIndicator.class);
    private final ShardingClient shardingClient;

    public ShardingHealthIndicator(ShardingClient shardingClient) {
        this.shardingClient = shardingClient;
    }

    @Override
    public Health health() {
        try {
            // Test connectivity by getting shard for a test key
            String testKey = "health-check-" + System.currentTimeMillis();
            shardingClient.getShardForKey(testKey);
            
            return Health.up()
                .withDetail("sharding-router", "Connected")
                .withDetail("status", "Operational")
                .build();
        } catch (ShardingClientException e) {
            log.error("Sharding system health check failed", e);
            return Health.down()
                .withDetail("sharding-router", "Unavailable")
                .withDetail("error", e.getMessage())
                .build();
        }
    }
}

