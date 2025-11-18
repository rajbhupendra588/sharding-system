package com.sharding.system.client.config;

import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.repository.ShardingRepositoryFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.HashMap;
import java.util.Map;

/**
 * Auto-configuration for Sharding Client.
 * Automatically creates and manages repository instances.
 * 
 * Usage:
 * <pre>
 * {@code
 * ShardingClientAutoConfiguration config = new ShardingClientAutoConfiguration();
 * config.setRouterUrl("http://localhost:8080");
 * config.scanRepositories("com.example.repository");
 * config.initialize();
 * 
 * UserRepository userRepo = config.getRepository(UserRepository.class);
 * }
 * </pre>
 */
public class ShardingClientAutoConfiguration {
    private static final Logger logger = LoggerFactory.getLogger(ShardingClientAutoConfiguration.class);
    
    private ShardingClient shardingClient;
    private String routerUrl = "http://localhost:8080";
    private final Map<Class<?>, Object> repositories = new HashMap<>();
    @SuppressWarnings("unused")
    private String basePackage; // Reserved for future package scanning feature
    
    /**
     * Sets the router URL.
     */
    public void setRouterUrl(String routerUrl) {
        this.routerUrl = routerUrl;
    }
    
    /**
     * Sets the base package to scan for repositories.
     */
    public void setBasePackage(String basePackage) {
        this.basePackage = basePackage;
    }
    
    /**
     * Initializes the configuration and creates the ShardingClient.
     */
    public void initialize() {
        if (shardingClient == null) {
            shardingClient = new ShardingClient(routerUrl);
            logger.info("ShardingClient initialized with router URL: {}", routerUrl);
        }
    }
    
    /**
     * Gets or creates a repository instance.
     */
    @SuppressWarnings("unchecked")
    public <T> T getRepository(Class<T> repositoryInterface) {
        if (shardingClient == null) {
            initialize();
        }
        
        return (T) repositories.computeIfAbsent(repositoryInterface, repo -> {
            logger.info("Creating repository instance for: {}", repositoryInterface.getName());
            return ShardingRepositoryFactory.createRepository(shardingClient, repositoryInterface);
        });
    }
    
    /**
     * Registers a repository manually.
     */
    public <T> void registerRepository(Class<T> repositoryInterface, T repository) {
        repositories.put(repositoryInterface, repository);
    }
    
    /**
     * Gets the ShardingClient instance.
     */
    public ShardingClient getShardingClient() {
        if (shardingClient == null) {
            initialize();
        }
        return shardingClient;
    }
    
    /**
     * Closes the configuration and releases resources.
     */
    public void close() {
        if (shardingClient != null) {
            try {
                shardingClient.close();
            } catch (Exception e) {
                logger.error("Error closing ShardingClient", e);
            }
        }
        repositories.clear();
    }
}

