package com.example.service;

import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.ShardingClientException;
import com.example.model.User;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import org.jboss.logging.Logger;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * User service demonstrating the correct sharding pattern.
 * 
 * KEY ARCHITECTURE POINTS:
 * 1. ShardingClient routes queries LOCALLY using cached shard map
 * 2. Queries are executed DIRECTLY on database shards (no proxy)
 * 3. No network call is made for routing decisions
 * 
 * This is how companies like Instagram, Pinterest, and Uber implement sharding.
 */
@ApplicationScoped
public class UserService {
    
    private static final Logger LOG = Logger.getLogger(UserService.class);
    
    @Inject
    ShardingClient shardingClient;
    
    /**
     * Get a user by ID.
     * 
     * The user_id is the SHARD KEY - it determines which database shard to query.
     * Routing happens LOCALLY in the ShardingClient using the cached shard map.
     * The query is then executed DIRECTLY on that database shard.
     */
    public User getUserById(String userId) throws ShardingClientException {
        LOG.debugf("Getting user: %s", userId);
        
        // This routes LOCALLY and queries DIRECTLY on the shard
        List<Map<String, Object>> rows = shardingClient.query(
            userId,  // Shard key - determines which DB shard to use
            "SELECT id, name, email FROM users WHERE id = ?",
            userId
        );
        
        if (rows.isEmpty()) {
            return null;
        }
        
        return mapRowToUser(rows.get(0));
    }
    
    /**
     * Create a new user.
     * 
     * The user is stored in the shard determined by their ID.
     */
    public void createUser(User user) throws ShardingClientException {
        LOG.infof("Creating user: %s", user.getId());
        
        shardingClient.update(
            user.getId(),  // Shard key
            "INSERT INTO users (id, name, email) VALUES (?, ?, ?)",
            user.getId(),
            user.getName(),
            user.getEmail()
        );
    }
    
    /**
     * Update an existing user.
     */
    public void updateUser(User user) throws ShardingClientException {
        LOG.infof("Updating user: %s", user.getId());
        
        shardingClient.update(
            user.getId(),  // Shard key
            "UPDATE users SET name = ?, email = ? WHERE id = ?",
            user.getName(),
            user.getEmail(),
            user.getId()
        );
    }
    
    /**
     * Delete a user.
     */
    public void deleteUser(String userId) throws ShardingClientException {
        LOG.infof("Deleting user: %s", userId);
        
        shardingClient.update(
            userId,  // Shard key
            "DELETE FROM users WHERE id = ?",
            userId
        );
    }
    
    /**
     * List users on a specific shard.
     * 
     * Note: This queries a single shard. To list ALL users across all shards,
     * use queryAllShards() which does scatter-gather (use sparingly!).
     */
    public List<User> listUsersOnShard(String anyKeyOnShard) throws ShardingClientException {
        LOG.debugf("Listing users for shard containing key: %s", anyKeyOnShard);
        
        List<Map<String, Object>> rows = shardingClient.query(
            anyKeyOnShard,
            "SELECT id, name, email FROM users ORDER BY id LIMIT 100"
        );
        
        List<User> users = new ArrayList<>();
        for (Map<String, Object> row : rows) {
            users.add(mapRowToUser(row));
        }
        
        return users;
    }
    
    /**
     * Get the shard ID for a user.
     * This is a LOCAL operation - no network call.
     */
    public String getShardForUser(String userId) {
        return shardingClient.getShardIdForKey(userId);
    }
    
    /**
     * Get shard statistics for monitoring.
     */
    public Map<String, ?> getShardStats() {
        return shardingClient.getPoolStats();
    }
    
    private User mapRowToUser(Map<String, Object> row) {
        User user = new User();
        user.setId((String) row.get("id"));
        user.setName((String) row.get("name"));
        user.setEmail((String) row.get("email"));
        return user;
    }
}
