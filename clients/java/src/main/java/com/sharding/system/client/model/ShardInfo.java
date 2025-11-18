package com.sharding.system.client.model;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * Response model for shard lookup.
 */
public class ShardInfo {
    @JsonProperty("shard_id")
    private String shardId;
    
    @JsonProperty("shard_name")
    private String shardName;
    
    @JsonProperty("hash_value")
    private Long hashValue;
    
    public String getShardId() {
        return shardId;
    }
    
    public void setShardId(String shardId) {
        this.shardId = shardId;
    }
    
    public String getShardName() {
        return shardName;
    }
    
    public void setShardName(String shardName) {
        this.shardName = shardName;
    }
    
    public Long getHashValue() {
        return hashValue;
    }
    
    public void setHashValue(Long hashValue) {
        this.hashValue = hashValue;
    }
}

