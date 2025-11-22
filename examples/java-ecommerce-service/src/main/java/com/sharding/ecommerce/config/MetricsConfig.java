package com.sharding.ecommerce.config;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * Metrics configuration for monitoring application performance.
 * 
 * Provides custom metrics for:
 * - Sharding operations
 * - API response times
 * - Error rates
 */
@Configuration
public class MetricsConfig {

    @Bean
    public Timer shardingQueryTimer(MeterRegistry registry) {
        return Timer.builder("sharding.query.duration")
            .description("Duration of sharding query operations")
            .register(registry);
    }

    @Bean
    public Timer shardingLookupTimer(MeterRegistry registry) {
        return Timer.builder("sharding.lookup.duration")
            .description("Duration of shard lookup operations")
            .register(registry);
    }
}

