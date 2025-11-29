package com.sharding.system.client.model;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * Represents a database shard configuration.
 * Contains all information needed to connect directly to the shard.
 */
public class Shard {
    
    @JsonProperty("id")
    private String id;
    
    @JsonProperty("name")
    private String name;
    
    @JsonProperty("endpoint")
    private String endpoint;
    
    @JsonProperty("status")
    private String status;
    
    @JsonProperty("hash_range_start")
    private long hashRangeStart;
    
    @JsonProperty("hash_range_end")
    private long hashRangeEnd;
    
    @JsonProperty("client_app_id")
    private String clientAppId;
    
    @JsonProperty("replica_endpoints")
    private String[] replicaEndpoints;
    
    public Shard() {
    }
    
    public Shard(String id, String endpoint, long hashRangeStart, long hashRangeEnd) {
        this.id = id;
        this.endpoint = endpoint;
        this.hashRangeStart = hashRangeStart;
        this.hashRangeEnd = hashRangeEnd;
        this.status = "active";
    }
    
    /**
     * Check if this shard owns the given hash value.
     */
    public boolean ownsHash(long hashValue) {
        // Handle wraparound case
        if (hashRangeStart <= hashRangeEnd) {
            return hashValue >= hashRangeStart && hashValue <= hashRangeEnd;
        } else {
            // Wraparound: e.g., start=3000000000, end=500000000
            return hashValue >= hashRangeStart || hashValue <= hashRangeEnd;
        }
    }
    
    public boolean isActive() {
        return "active".equals(status);
    }
    
    // Getters and Setters
    
    public String getId() {
        return id;
    }
    
    public void setId(String id) {
        this.id = id;
    }
    
    public String getName() {
        return name;
    }
    
    public void setName(String name) {
        this.name = name;
    }
    
    public String getEndpoint() {
        return endpoint;
    }
    
    public void setEndpoint(String endpoint) {
        this.endpoint = endpoint;
    }
    
    public String getStatus() {
        return status;
    }
    
    public void setStatus(String status) {
        this.status = status;
    }
    
    public long getHashRangeStart() {
        return hashRangeStart;
    }
    
    public void setHashRangeStart(long hashRangeStart) {
        this.hashRangeStart = hashRangeStart;
    }
    
    public long getHashRangeEnd() {
        return hashRangeEnd;
    }
    
    public void setHashRangeEnd(long hashRangeEnd) {
        this.hashRangeEnd = hashRangeEnd;
    }
    
    public String getClientAppId() {
        return clientAppId;
    }
    
    public void setClientAppId(String clientAppId) {
        this.clientAppId = clientAppId;
    }
    
    public String[] getReplicaEndpoints() {
        return replicaEndpoints;
    }
    
    public void setReplicaEndpoints(String[] replicaEndpoints) {
        this.replicaEndpoints = replicaEndpoints;
    }
    
    @Override
    public String toString() {
        return "Shard{id='" + id + "', endpoint='" + endpoint + "', range=[" + hashRangeStart + "," + hashRangeEnd + "]}";
    }
}

