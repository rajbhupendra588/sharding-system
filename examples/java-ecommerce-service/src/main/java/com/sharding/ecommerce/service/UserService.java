package com.sharding.ecommerce.service;

import com.sharding.ecommerce.model.User;
import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.ShardingClientException;
import com.sharding.system.client.model.QueryResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;
import java.util.UUID;

/**
 * User Service demonstrating sharding by user ID.
 * 
 * All user operations use the user ID as the shard key, ensuring:
 * - User data is distributed evenly across shards
 * - User-related queries are fast (no cross-shard operations)
 * - User and their orders are co-located on the same shard
 */
@Service
public class UserService {

    private static final Logger log = LoggerFactory.getLogger(UserService.class);
    private final ShardingClient shardingClient;

    public UserService(ShardingClient shardingClient) {
        this.shardingClient = shardingClient;
    }

    /**
     * Creates a new user.
     * The user ID is used as the shard key for routing.
     */
    public User createUser(User user) throws ShardingClientException {
        log.info("Creating user with ID: {}", user.getId());
        
        String userId = user.getId() != null ? user.getId() : UUID.randomUUID().toString();
        user.setId(userId);
        user.setCreatedAt(LocalDateTime.now());
        user.setUpdatedAt(LocalDateTime.now());
        
        // Use user ID as shard key - ensures user data goes to the correct shard
        shardingClient.queryStrong(
            userId,
            """
                INSERT INTO users (id, username, email, full_name, phone_number, address, created_at, updated_at, active)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            """,
            userId,
            user.getUsername(),
            user.getEmail(),
            user.getFullName(),
            user.getPhoneNumber(),
            user.getAddress(),
            user.getCreatedAt(),
            user.getUpdatedAt(),
            user.getActive() != null ? user.getActive() : true
        );
        
        log.info("User created successfully. Shard key: {}", userId);
        return user;
    }

    /**
     * Retrieves a user by ID.
     * Uses strong consistency to ensure we get the latest data.
     */
    public User getUserById(String userId) throws ShardingClientException {
        log.debug("Fetching user with ID: {}", userId);
        
        QueryResponse response = shardingClient.queryStrong(
            userId, // Shard key
            "SELECT id, username, email, full_name, phone_number, address, created_at, updated_at, active FROM users WHERE id = $1",
            userId
        );
        
        if (response.getRowCount() == 0) {
            log.warn("User not found: {}", userId);
            return null;
        }
        
        return mapRowToUser(response.getRows().get(0));
    }

    /**
     * Updates user information.
     * Uses the user ID as shard key for routing.
     */
    public User updateUser(String userId, User user) throws ShardingClientException {
        log.info("Updating user with ID: {}", userId);
        
        user.setUpdatedAt(LocalDateTime.now());
        
        shardingClient.queryStrong(
            userId, // Shard key
            """
                UPDATE users 
                SET username = $2, email = $3, full_name = $4, phone_number = $5, 
                    address = $6, updated_at = $7, active = $8
                WHERE id = $1
            """,
            userId,
            user.getUsername(),
            user.getEmail(),
            user.getFullName(),
            user.getPhoneNumber(),
            user.getAddress(),
            user.getUpdatedAt(),
            user.getActive() != null ? user.getActive() : true
        );
        
        log.info("User updated successfully. Shard key: {}", userId);
        return getUserById(userId);
    }

    /**
     * Deletes a user (soft delete by setting active = false).
     */
    public void deleteUser(String userId) throws ShardingClientException {
        log.info("Deleting user with ID: {}", userId);
        
        shardingClient.queryStrong(
            userId, // Shard key
            "UPDATE users SET active = false, updated_at = $2 WHERE id = $1",
            userId,
            LocalDateTime.now()
        );
        
        log.info("User deleted successfully. Shard key: {}", userId);
    }

    /**
     * Gets shard information for a user ID.
     * Useful for debugging and understanding shard distribution.
     */
    public String getShardForUser(String userId) throws ShardingClientException {
        log.debug("Getting shard for user ID: {}", userId);
        String shardId = shardingClient.getShardForKey(userId);
        log.debug("User {} is on shard: {}", userId, shardId);
        return shardId;
    }

    private User mapRowToUser(Map<String, Object> row) {
        return User.builder()
            .id((String) row.get("id"))
            .username((String) row.get("username"))
            .email((String) row.get("email"))
            .fullName((String) row.get("full_name"))
            .phoneNumber((String) row.get("phone_number"))
            .address((String) row.get("address"))
            .createdAt((LocalDateTime) row.get("created_at"))
            .updatedAt((LocalDateTime) row.get("updated_at"))
            .active((Boolean) row.get("active"))
            .build();
    }
}

