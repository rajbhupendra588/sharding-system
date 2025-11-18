package com.example.repository;

import com.example.model.UserEntity;
import com.sharding.system.client.annotation.EventualConsistency;
import com.sharding.system.client.annotation.Query;
import com.sharding.system.client.annotation.ShardingRepository;
import com.sharding.system.client.annotation.ShardKey;
import com.sharding.system.client.repository.CrudRepository;

import java.util.List;
import java.util.Optional;

/**
 * User repository - ZERO implementation code!
 * All methods are automatically implemented by the framework.
 */
@ShardingRepository(entity = UserEntity.class, table = "users")
public interface UserRepository extends CrudRepository<UserEntity, String> {
    
    // Auto-generated: SELECT * FROM users WHERE email = $1
    Optional<UserEntity> findByEmail(String email);
    
    // Custom query with automatic mapping
    @Query("SELECT * FROM users WHERE name LIKE $1 ORDER BY name LIMIT $2")
    List<UserEntity> findByNameLike(String namePattern, int limit);
    
    // Eventual consistency for read-heavy operations
    @EventualConsistency
    @Query("SELECT * FROM users WHERE status = $1")
    List<UserEntity> findByStatus(String status);
    
    // Complex query with multiple conditions
    @Query("SELECT * FROM users WHERE email = $1 AND status = $2")
    Optional<UserEntity> findByEmailAndStatus(String email, String status);
}

