package com.example.service;

import com.sharding.system.client.ShardingClient;
import io.quarkus.arc.DefaultBean;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.enterprise.inject.Produces;
import org.eclipse.microprofile.config.inject.ConfigProperty;

@ApplicationScoped
public class ShardingClientProducer {
    
    @ConfigProperty(name = "sharding.router.url", defaultValue = "http://localhost:8080")
    String routerUrl;
    
    @Produces
    @ApplicationScoped
    @DefaultBean
    public ShardingClient shardingClient() {
        return new ShardingClient(routerUrl);
    }
}

