package com.sharding.system.client;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.sharding.system.client.model.QueryRequest;
import com.sharding.system.client.model.QueryResponse;
import com.sharding.system.client.model.ShardInfo;
import org.apache.hc.client5.http.classic.methods.HttpGet;
import org.apache.hc.client5.http.classic.methods.HttpPost;
import org.apache.hc.client5.http.impl.classic.CloseableHttpClient;
import org.apache.hc.client5.http.impl.classic.CloseableHttpResponse;
import org.apache.hc.client5.http.impl.classic.HttpClients;
import org.apache.hc.core5.http.ClassicHttpResponse;
import org.apache.hc.core5.http.io.entity.EntityUtils;
import org.apache.hc.core5.http.io.entity.StringEntity;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.List;
import org.apache.hc.core5.http.ParseException;

/**
 * Java client for Sharding System Router API.
 * 
 * This client provides a simple interface for executing queries against sharded databases
 * and looking up shard information for keys.
 */
public class ShardingClient {
    private static final Logger logger = LoggerFactory.getLogger(ShardingClient.class);
    
    private final String routerUrl;
    private final CloseableHttpClient httpClient;
    private final ObjectMapper objectMapper;
    
    /**
     * Creates a new ShardingClient instance.
     * 
     * @param routerUrl Base URL of the sharding router (e.g., "http://localhost:8080")
     */
    public ShardingClient(String routerUrl) {
        this.routerUrl = routerUrl.endsWith("/") ? routerUrl.substring(0, routerUrl.length() - 1) : routerUrl;
        this.httpClient = HttpClients.createDefault();
        this.objectMapper = new ObjectMapper();
    }
    
    /**
     * Gets the shard ID for a given key.
     * 
     * @param key The shard key
     * @return The shard ID that owns this key
     * @throws ShardingClientException if the request fails
     */
    public String getShardForKey(String key) throws ShardingClientException {
        try {
            String encodedKey = URLEncoder.encode(key, StandardCharsets.UTF_8.toString());
            String url = routerUrl + "/v1/shard-for-key?key=" + encodedKey;
            
            HttpGet request = new HttpGet(url);
            
            ClassicHttpResponse response = httpClient.executeOpen(null, request, null);
            try {
                int statusCode = response.getCode();
                String responseBody = EntityUtils.toString(response.getEntity());
                
                if (statusCode != 200) {
                    throw new ShardingClientException("Failed to get shard for key: " + key + 
                                                     ". Status: " + statusCode + ", Response: " + responseBody);
                }
                
                ShardInfo shardInfo = objectMapper.readValue(responseBody, ShardInfo.class);
                return shardInfo.getShardId();
            } finally {
                if (response instanceof CloseableHttpResponse) {
                    ((CloseableHttpResponse) response).close();
                }
            }
        } catch (IOException | ParseException e) {
            logger.error("Error getting shard for key: {}", key, e);
            throw new ShardingClientException("Failed to get shard for key: " + key, e);
        }
    }
    
    /**
     * Executes a query with the specified consistency level.
     * 
     * @param shardKey The shard key for routing
     * @param query SQL query string
     * @param params Query parameters
     * @param consistency Consistency level ("strong" or "eventual")
     * @return QueryResponse containing results
     * @throws ShardingClientException if the query fails
     */
    public QueryResponse query(String shardKey, String query, List<Object> params, String consistency) 
            throws ShardingClientException {
        try {
            QueryRequest request = new QueryRequest();
            request.setShardKey(shardKey);
            request.setQuery(query);
            request.setParams(params);
            request.setConsistency(consistency != null ? consistency : "strong");
            
            String url = routerUrl + "/v1/execute";
            String requestBody = objectMapper.writeValueAsString(request);
            
            HttpPost httpPost = new HttpPost(url);
            httpPost.setHeader("Content-Type", "application/json");
            httpPost.setEntity(new StringEntity(requestBody, StandardCharsets.UTF_8));
            
            ClassicHttpResponse response = httpClient.executeOpen(null, httpPost, null);
            try {
                int statusCode = response.getCode();
                String responseBody = EntityUtils.toString(response.getEntity());
                
                if (statusCode != 200) {
                    throw new ShardingClientException("Query execution failed. Status: " + statusCode + 
                                                     ", Response: " + responseBody);
                }
                
                return objectMapper.readValue(responseBody, QueryResponse.class);
            } finally {
                if (response instanceof CloseableHttpResponse) {
                    ((CloseableHttpResponse) response).close();
                }
            }
        } catch (IOException | ParseException e) {
            logger.error("Error executing query: {}", query, e);
            throw new ShardingClientException("Failed to execute query: " + query, e);
        }
    }
    
    /**
     * Executes a query with strong consistency (reads from primary).
     * 
     * @param shardKey The shard key for routing
     * @param query SQL query string
     * @param params Query parameters
     * @return QueryResponse containing results
     * @throws ShardingClientException if the query fails
     */
    public QueryResponse queryStrong(String shardKey, String query, List<Object> params) 
            throws ShardingClientException {
        return query(shardKey, query, params, "strong");
    }
    
    /**
     * Executes a query with eventual consistency (can read from replica).
     * 
     * @param shardKey The shard key for routing
     * @param query SQL query string
     * @param params Query parameters
     * @return QueryResponse containing results
     * @throws ShardingClientException if the query fails
     */
    public QueryResponse queryEventual(String shardKey, String query, List<Object> params) 
            throws ShardingClientException {
        return query(shardKey, query, params, "eventual");
    }
    
    /**
     * Executes a query with strong consistency (convenience method with varargs).
     * 
     * @param shardKey The shard key for routing
     * @param query SQL query string
     * @param params Query parameters (varargs)
     * @return QueryResponse containing results
     * @throws ShardingClientException if the query fails
     */
    public QueryResponse queryStrong(String shardKey, String query, Object... params) 
            throws ShardingClientException {
        return queryStrong(shardKey, query, List.of(params));
    }
    
    /**
     * Executes a query with eventual consistency (convenience method with varargs).
     * 
     * @param shardKey The shard key for routing
     * @param query SQL query string
     * @param params Query parameters (varargs)
     * @return QueryResponse containing results
     * @throws ShardingClientException if the query fails
     */
    public QueryResponse queryEventual(String shardKey, String query, Object... params) 
            throws ShardingClientException {
        return queryEventual(shardKey, query, List.of(params));
    }
    
    /**
     * Closes the HTTP client and releases resources.
     */
    public void close() throws IOException {
        httpClient.close();
    }
}

