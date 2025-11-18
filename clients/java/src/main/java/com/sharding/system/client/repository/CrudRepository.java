package com.sharding.system.client.repository;

import com.sharding.system.client.annotation.ShardKey;
import java.util.List;
import java.util.Optional;

/**
 * Base repository interface providing CRUD operations.
 * Extend this interface in your repository to get automatic CRUD functionality.
 * 
 * @param <T> The entity type
 * @param <ID> The ID type (usually String)
 */
public interface CrudRepository<T, ID> {
    
    /**
     * Saves an entity. Performs INSERT if new, UPDATE if exists.
     * Automatically extracts shard key from entity.
     */
    T save(T entity);
    
    /**
     * Saves all entities in a batch.
     */
    List<T> saveAll(Iterable<T> entities);
    
    /**
     * Finds an entity by ID.
     * Automatically extracts shard key from ID.
     */
    Optional<T> findById(@ShardKey ID id);
    
    /**
     * Checks if an entity exists by ID.
     */
    boolean existsById(@ShardKey ID id);
    
    /**
     * Finds all entities. Requires a shard key for routing.
     */
    List<T> findAll(@ShardKey String shardKey);
    
    /**
     * Counts all entities in a shard.
     */
    long count(@ShardKey String shardKey);
    
    /**
     * Deletes an entity by ID.
     */
    void deleteById(@ShardKey ID id);
    
    /**
     * Deletes an entity.
     */
    void delete(T entity);
    
    /**
     * Deletes all entities in a shard.
     */
    void deleteAll(@ShardKey String shardKey);
}

