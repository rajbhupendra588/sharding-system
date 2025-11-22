package com.sharding.ecommerce.config;

import com.sharding.system.client.ShardingClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;

/**
 * Configuration for Sharding Client.
 * 
 * This configuration creates a singleton ShardingClient bean that will be used
 * throughout the application to interact with the sharding router.
 */
@Configuration
public class ShardingConfig {

    @Value("${sharding.router.url}")
    private String routerUrl;

    @Bean
    @Primary
    public ShardingClient shardingClient() {
        return new ShardingClient(routerUrl);
    }
}

