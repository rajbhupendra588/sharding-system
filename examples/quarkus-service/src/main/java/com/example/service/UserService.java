package com.example.service;

import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.ShardingClientException;
import com.sharding.system.client.model.QueryResponse;
import com.example.model.User;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import org.jboss.logging.Logger;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

@ApplicationScoped
public class UserService {
    
    private static final Logger LOG = Logger.getLogger(UserService.class);
    
    @Inject
    ShardingClient shardingClient;
    
    public User getUserById(String userId) throws ShardingClientException {
        LOG.infof("Getting user: %s", userId);
        
        QueryResponse response = shardingClient.queryStrong(
            userId,
            "SELECT id, name, email FROM users WHERE id = $1",
            userId
        );
        
        if (response.getRowCount() == 0) {
            return null;
        }
        
        return mapRowToUser(response.getRows().get(0));
    }
    
    public void createUser(User user) throws ShardingClientException {
        LOG.infof("Creating user: %s", user.getId());
        
        shardingClient.queryStrong(
            user.getId(),
            "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)",
            user.getId(),
            user.getName(),
            user.getEmail()
        );
    }
    
    public void updateUser(User user) throws ShardingClientException {
        LOG.infof("Updating user: %s", user.getId());
        
        shardingClient.queryStrong(
            user.getId(),
            "UPDATE users SET name = $1, email = $2 WHERE id = $3",
            user.getName(),
            user.getEmail(),
            user.getId()
        );
    }
    
    public void deleteUser(String userId) throws ShardingClientException {
        LOG.infof("Deleting user: %s", userId);
        
        shardingClient.queryStrong(
            userId,
            "DELETE FROM users WHERE id = $1",
            userId
        );
    }
    
    public List<User> listUsers(String shardKey) throws ShardingClientException {
        LOG.infof("Listing users for shard key: %s", shardKey);
        
        QueryResponse response = shardingClient.queryEventual(
            shardKey,
            "SELECT id, name, email FROM users ORDER BY id LIMIT 100"
        );
        
        List<User> users = new ArrayList<>();
        for (Map<String, Object> row : response.getRows()) {
            users.add(mapRowToUser(row));
        }
        
        return users;
    }
    
    public String getShardForKey(String key) throws ShardingClientException {
        return shardingClient.getShardForKey(key);
    }
    
    private User mapRowToUser(Map<String, Object> row) {
        User user = new User();
        user.setId((String) row.get("id"));
        user.setName((String) row.get("name"));
        user.setEmail((String) row.get("email"));
        return user;
    }
}

