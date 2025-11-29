package com.example.service;

import com.sharding.system.client.ShardingClient;
import io.quarkus.arc.DefaultBean;
import io.quarkus.runtime.Shutdown;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.enterprise.inject.Produces;
import org.eclipse.microprofile.config.inject.ConfigProperty;
import org.jboss.logging.Logger;

/**
 * CDI Producer for ShardingClient.
 * 
 * Creates a ShardingClient that:
 * 1. Fetches shard configuration from Manager on startup
 * 2. Caches shard map locally
 * 3. Maintains connection pools to all database shards
 * 4. Routes queries LOCALLY (no network call for routing)
 * 5. Executes queries DIRECTLY on database shards
 */
@ApplicationScoped
public class ShardingClientProducer {
    
    private static final Logger LOG = Logger.getLogger(ShardingClientProducer.class);
    
    @ConfigProperty(name = "sharding.manager.url", defaultValue = "http://localhost:8081")
    String managerUrl;
    
    @ConfigProperty(name = "sharding.client.app.id")
    String clientAppId;
    
    @ConfigProperty(name = "sharding.refresh.interval.seconds", defaultValue = "60")
    long refreshInterval;
    
    @ConfigProperty(name = "sharding.pool.max.size", defaultValue = "10")
    int maxPoolSize;
    
    @ConfigProperty(name = "sharding.pool.min.idle", defaultValue = "2")
    int minIdle;
    
    private ShardingClient shardingClient;
    
    @Produces
    @ApplicationScoped
    @DefaultBean
    public ShardingClient shardingClient() {
        LOG.infof("Initializing ShardingClient for app: %s", clientAppId);
        LOG.infof("Manager URL: %s", managerUrl);
        
        this.shardingClient = new ShardingClient.Builder()
            .managerUrl(managerUrl)
            .clientAppId(clientAppId)
            .refreshInterval(refreshInterval)
            .maxPoolSize(maxPoolSize)
            .minIdle(minIdle)
            .build();
        
        LOG.infof("ShardingClient initialized with %d active shards", 
            shardingClient.getShardMap().getActiveShards().size());
        
        return shardingClient;
    }
    
    @Shutdown
    void shutdown() {
        if (shardingClient != null) {
            LOG.info("Shutting down ShardingClient...");
            shardingClient.close();
        }
    }
}
