package com.sharding.system.client.model;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * Represents the shard map configuration.
 * Contains all shards for a client application and provides routing logic.
 * 
 * This is cached locally in the application and updated only when shards change.
 * All routing decisions are made locally - NO network calls during query routing.
 */
public class ShardMap {
    
    @JsonProperty("client_app_id")
    private String clientAppId;
    
    @JsonProperty("client_app_name")
    private String clientAppName;
    
    @JsonProperty("shards")
    private List<Shard> shards;
    
    @JsonProperty("version")
    private long version;
    
    // Thread-safe list for concurrent access
    private transient CopyOnWriteArrayList<Shard> activeShards;
    
    public ShardMap() {
        this.shards = new ArrayList<>();
    }
    
    public ShardMap(String clientAppId, List<Shard> shards) {
        this.clientAppId = clientAppId;
        this.shards = shards;
        this.version = System.currentTimeMillis();
        refreshActiveShards();
    }
    
    /**
     * Get the shard that owns the given key.
     * This is a LOCAL operation - no network call.
     * 
     * @param shardKey The key to route (e.g., user_id, order_id)
     * @return The shard that owns this key
     * @throws IllegalStateException if no shard found for the key
     */
    public Shard getShardForKey(String shardKey) {
        if (activeShards == null || activeShards.isEmpty()) {
            refreshActiveShards();
        }
        
        if (activeShards.isEmpty()) {
            throw new IllegalStateException("No active shards available");
        }
        
        // Calculate hash of the key
        long hashValue = hash(shardKey);
        
        // Find the shard that owns this hash
        for (Shard shard : activeShards) {
            if (shard.ownsHash(hashValue)) {
                return shard;
            }
        }
        
        // Fallback: if no range match, use modulo (shouldn't happen with proper config)
        int index = (int) (Math.abs(hashValue) % activeShards.size());
        return activeShards.get(index);
    }
    
    /**
     * Get the shard ID for a given key.
     * Convenience method that returns just the shard ID.
     */
    public String getShardIdForKey(String shardKey) {
        return getShardForKey(shardKey).getId();
    }
    
    /**
     * Calculate hash value for a shard key using MurmurHash3-like algorithm.
     * This matches the Go implementation in pkg/hashing/hashing.go
     */
    public static long hash(String key) {
        try {
            MessageDigest md = MessageDigest.getInstance("MD5");
            byte[] digest = md.digest(key.getBytes(StandardCharsets.UTF_8));
            
            // Convert first 8 bytes to long (matching Go's binary.BigEndian.Uint64)
            long hash = 0;
            for (int i = 0; i < 8; i++) {
                hash = (hash << 8) | (digest[i] & 0xFF);
            }
            
            // Return absolute value to ensure positive
            return Math.abs(hash);
        } catch (NoSuchAlgorithmException e) {
            // Fallback to simple hash
            return Math.abs(key.hashCode());
        }
    }
    
    /**
     * Refresh the list of active shards.
     * Called when shard map is updated.
     */
    public void refreshActiveShards() {
        List<Shard> active = new ArrayList<>();
        if (shards != null) {
            for (Shard shard : shards) {
                if (shard.isActive()) {
                    active.add(shard);
                }
            }
        }
        this.activeShards = new CopyOnWriteArrayList<>(active);
    }
    
    /**
     * Get all shards (including inactive).
     */
    public List<Shard> getAllShards() {
        return shards != null ? new ArrayList<>(shards) : new ArrayList<>();
    }
    
    /**
     * Get only active shards.
     */
    public List<Shard> getActiveShards() {
        if (activeShards == null) {
            refreshActiveShards();
        }
        return new ArrayList<>(activeShards);
    }
    
    /**
     * Get a shard by ID.
     */
    public Shard getShardById(String shardId) {
        if (shards == null) return null;
        for (Shard shard : shards) {
            if (shard.getId().equals(shardId)) {
                return shard;
            }
        }
        return null;
    }
    
    /**
     * Check if the shard map has any active shards.
     */
    public boolean hasActiveShards() {
        if (activeShards == null) {
            refreshActiveShards();
        }
        return !activeShards.isEmpty();
    }
    
    // Getters and Setters
    
    public String getClientAppId() {
        return clientAppId;
    }
    
    public void setClientAppId(String clientAppId) {
        this.clientAppId = clientAppId;
    }
    
    public String getClientAppName() {
        return clientAppName;
    }
    
    public void setClientAppName(String clientAppName) {
        this.clientAppName = clientAppName;
    }
    
    public List<Shard> getShards() {
        return shards;
    }
    
    public void setShards(List<Shard> shards) {
        this.shards = shards;
        refreshActiveShards();
    }
    
    public long getVersion() {
        return version;
    }
    
    public void setVersion(long version) {
        this.version = version;
    }
    
    @Override
    public String toString() {
        return "ShardMap{clientAppId='" + clientAppId + "', shards=" + (shards != null ? shards.size() : 0) + ", version=" + version + "}";
    }
}

