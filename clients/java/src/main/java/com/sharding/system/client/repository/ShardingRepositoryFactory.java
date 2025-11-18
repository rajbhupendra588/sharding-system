package com.sharding.system.client.repository;

import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.annotation.ShardingRepository;

import java.lang.reflect.Proxy;

/**
 * Factory for creating repository instances.
 * Creates dynamic proxies that implement repository interfaces.
 */
public class ShardingRepositoryFactory {
    
    /**
     * Creates a repository instance for the given interface.
     * 
     * @param shardingClient The sharding client to use
     * @param repositoryInterface The repository interface class
     * @return A proxy instance implementing the repository interface
     */
    @SuppressWarnings("unchecked")
    public static <T> T createRepository(ShardingClient shardingClient, Class<T> repositoryInterface) {
        if (!repositoryInterface.isInterface()) {
            throw new IllegalArgumentException("Repository must be an interface: " + repositoryInterface.getName());
        }
        
        ShardingRepository annotation = repositoryInterface.getAnnotation(ShardingRepository.class);
        if (annotation == null) {
            throw new IllegalArgumentException("Repository interface must be annotated with @ShardingRepository: " + repositoryInterface.getName());
        }
        
        Class<?> entityClass = annotation.entity();
        if (entityClass == null) {
            throw new IllegalArgumentException("Entity class must be specified in @ShardingRepository annotation");
        }
        
        ShardingRepositoryProxy proxy = new ShardingRepositoryProxy(shardingClient, repositoryInterface, entityClass);
        
        return (T) Proxy.newProxyInstance(
            repositoryInterface.getClassLoader(),
            new Class[]{repositoryInterface},
            proxy
        );
    }
}

