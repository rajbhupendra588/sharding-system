package com.sharding.system.client.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;
import java.util.Map;

/**
 * Response model for query execution.
 */
public class QueryResponse {
    @JsonProperty("shard_id")
    private String shardId;
    
    @JsonProperty("rows")
    private List<Map<String, Object>> rows;
    
    @JsonProperty("row_count")
    private int rowCount;
    
    @JsonProperty("latency_ms")
    private double latencyMs;
    
    public String getShardId() {
        return shardId;
    }
    
    public void setShardId(String shardId) {
        this.shardId = shardId;
    }
    
    public List<Map<String, Object>> getRows() {
        return rows;
    }
    
    public void setRows(List<Map<String, Object>> rows) {
        this.rows = rows;
    }
    
    public int getRowCount() {
        return rowCount;
    }
    
    public void setRowCount(int rowCount) {
        this.rowCount = rowCount;
    }
    
    public double getLatencyMs() {
        return latencyMs;
    }
    
    public void setLatencyMs(double latencyMs) {
        this.latencyMs = latencyMs;
    }
}

