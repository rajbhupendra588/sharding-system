package com.sharding.system.client.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;
import java.util.Map;

/**
 * Request model for query execution.
 */
public class QueryRequest {
    @JsonProperty("shard_key")
    private String shardKey;
    
    @JsonProperty("query")
    private String query;
    
    @JsonProperty("params")
    private List<Object> params;
    
    @JsonProperty("consistency")
    private String consistency;
    
    @JsonProperty("options")
    private Map<String, Object> options;
    
    public String getShardKey() {
        return shardKey;
    }
    
    public void setShardKey(String shardKey) {
        this.shardKey = shardKey;
    }
    
    public String getQuery() {
        return query;
    }
    
    public void setQuery(String query) {
        this.query = query;
    }
    
    public List<Object> getParams() {
        return params;
    }
    
    public void setParams(List<Object> params) {
        this.params = params;
    }
    
    public String getConsistency() {
        return consistency;
    }
    
    public void setConsistency(String consistency) {
        this.consistency = consistency;
    }
    
    public Map<String, Object> getOptions() {
        return options;
    }
    
    public void setOptions(Map<String, Object> options) {
        this.options = options;
    }
}

