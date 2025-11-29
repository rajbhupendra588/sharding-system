package com.sharding.system.client;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sharding.system.client.model.Shard;
import com.sharding.system.client.model.ShardMap;
import org.apache.hc.client5.http.classic.methods.HttpGet;
import org.apache.hc.client5.http.impl.classic.CloseableHttpClient;
import org.apache.hc.client5.http.impl.classic.CloseableHttpResponse;
import org.apache.hc.client5.http.impl.classic.HttpClients;
import org.apache.hc.core5.http.io.entity.EntityUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.sql.*;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

/**
 * High-performance Java client for Sharding System.
 * 
 * <h2>Architecture</h2>
 * This client implements the CORRECT sharding pattern used by companies like
 * Instagram, Pinterest, and Uber:
 * 
 * <pre>
 * ┌─────────────────────────────────────────────────────────────────┐
 * │                     CONTROL PLANE (startup only)                │
 * │  Manager API ──► Shard Map (cached locally)                     │
 * └─────────────────────────────────────────────────────────────────┘
 *                              │
 *                              ▼ (cached, no network call)
 * ┌─────────────────────────────────────────────────────────────────┐
 * │                     DATA PLANE (every query)                    │
 * │  ShardingClient ──► Connection Pool ──► Database Shard DIRECTLY │
 * └─────────────────────────────────────────────────────────────────┘
 * </pre>
 * 
 * <h2>Key Features</h2>
 * <ul>
 *   <li>Caches shard map locally - routing decisions have ZERO network latency</li>
 *   <li>Maintains connection pools to ALL shards for direct database access</li>
 *   <li>No proxy/router in the query path - maximum performance</li>
 *   <li>Automatic shard map refresh when configuration changes</li>
 * </ul>
 * 
 * <h2>Usage</h2>
 * <pre>
 * ShardingClient client = new ShardingClient.Builder()
 *     .managerUrl("http://localhost:8081")
 *     .clientAppId("your-app-id")
 *     .build();
 * 
 * // Queries are routed based on shard key and executed DIRECTLY on the database
 * List&lt;Map&gt; users = client.query("user-123", "SELECT * FROM users WHERE id = ?", "user-123");
 * </pre>
 */
public class ShardingClient implements AutoCloseable {
    
    private static final Logger logger = LoggerFactory.getLogger(ShardingClient.class);
    
    // Configuration
    private final String managerUrl;
    private final String clientAppId;
    
    // Cached shard map - updated only when shards change
    private final AtomicReference<ShardMap> shardMap;
    
    // Connection pools - one per shard, for DIRECT database access
    private final ShardConnectionManager connectionManager;
    
    // HTTP client for fetching shard config from Manager
    private final CloseableHttpClient httpClient;
    private final ObjectMapper objectMapper;
    
    // Background refresh
    private final ScheduledExecutorService refreshExecutor;
    private final long refreshIntervalSeconds;
    
    /**
     * Creates a new ShardingClient using the Builder.
     */
    private ShardingClient(Builder builder) {
        this.managerUrl = builder.managerUrl.endsWith("/") 
            ? builder.managerUrl.substring(0, builder.managerUrl.length() - 1) 
            : builder.managerUrl;
        this.clientAppId = builder.clientAppId;
        this.refreshIntervalSeconds = builder.refreshIntervalSeconds;
        
        this.shardMap = new AtomicReference<>();
        this.connectionManager = new ShardConnectionManager(
            builder.maxPoolSize,
            builder.minIdle,
            builder.connectionTimeout,
            builder.idleTimeout,
            builder.maxLifetime
        );
        this.httpClient = HttpClients.createDefault();
        this.objectMapper = new ObjectMapper();
        
        // Initialize shard map from Manager
        try {
            refreshShardMap();
        } catch (ShardingClientException e) {
            logger.error("Failed to initialize shard map", e);
            throw new RuntimeException("Failed to initialize ShardingClient: " + e.getMessage(), e);
        }
        
        // Start background refresh
        this.refreshExecutor = Executors.newSingleThreadScheduledExecutor(r -> {
            Thread t = new Thread(r, "shard-map-refresh");
            t.setDaemon(true);
            return t;
        });
        
        refreshExecutor.scheduleAtFixedRate(
            this::refreshShardMapSilent,
            refreshIntervalSeconds,
            refreshIntervalSeconds,
            TimeUnit.SECONDS
        );
        
        logger.info("ShardingClient initialized for app: {} with {} shards", 
            clientAppId, shardMap.get().getActiveShards().size());
    }
    
    /**
     * Execute a query with automatic shard routing.
     * 
     * The shard key determines which database shard to query.
     * Routing is done LOCALLY using the cached shard map - no network call.
     * Query is executed DIRECTLY on the database - no proxy.
     * 
     * @param shardKey The key used for routing (e.g., user_id, order_id)
     * @param sql The SQL query with ? placeholders
     * @param params Query parameters
     * @return List of rows as Maps
     * @throws ShardingClientException if query fails
     */
    public List<Map<String, Object>> query(String shardKey, String sql, Object... params) 
            throws ShardingClientException {
        
        // 1. Route to shard LOCALLY (no network call)
        Shard shard = shardMap.get().getShardForKey(shardKey);
        
        logger.debug("Routing key '{}' to shard: {}", shardKey, shard.getId());
        
        // 2. Execute query DIRECTLY on the database
        try (Connection conn = connectionManager.getConnection(shard.getId());
             PreparedStatement ps = conn.prepareStatement(sql)) {
            
            // Set parameters
            for (int i = 0; i < params.length; i++) {
                ps.setObject(i + 1, params[i]);
            }
            
            // Execute and map results
            try (ResultSet rs = ps.executeQuery()) {
                return mapResultSet(rs);
            }
            
        } catch (SQLException e) {
            logger.error("Query failed on shard {}: {}", shard.getId(), sql, e);
            throw new ShardingClientException("Query failed: " + e.getMessage(), e);
        }
    }
    
    /**
     * Execute an update (INSERT, UPDATE, DELETE) with automatic shard routing.
     * 
     * @param shardKey The key used for routing
     * @param sql The SQL statement with ? placeholders
     * @param params Statement parameters
     * @return Number of affected rows
     * @throws ShardingClientException if update fails
     */
    public int update(String shardKey, String sql, Object... params) throws ShardingClientException {
        
        // 1. Route to shard LOCALLY
        Shard shard = shardMap.get().getShardForKey(shardKey);
        
        logger.debug("Routing update for key '{}' to shard: {}", shardKey, shard.getId());
        
        // 2. Execute DIRECTLY on the database
        try (Connection conn = connectionManager.getConnection(shard.getId());
             PreparedStatement ps = conn.prepareStatement(sql)) {
            
            // Set parameters
            for (int i = 0; i < params.length; i++) {
                ps.setObject(i + 1, params[i]);
            }
            
            return ps.executeUpdate();
            
        } catch (SQLException e) {
            logger.error("Update failed on shard {}: {}", shard.getId(), sql, e);
            throw new ShardingClientException("Update failed: " + e.getMessage(), e);
        }
    }
    
    /**
     * Execute a query on a specific shard (bypass routing).
     * Useful for admin operations or cross-shard queries.
     * 
     * @param shardId The specific shard ID
     * @param sql The SQL query
     * @param params Query parameters
     * @return List of rows as Maps
     */
    public List<Map<String, Object>> queryOnShard(String shardId, String sql, Object... params) 
            throws ShardingClientException {
        
        try (Connection conn = connectionManager.getConnection(shardId);
             PreparedStatement ps = conn.prepareStatement(sql)) {
            
            for (int i = 0; i < params.length; i++) {
                ps.setObject(i + 1, params[i]);
            }
            
            try (ResultSet rs = ps.executeQuery()) {
                return mapResultSet(rs);
            }
            
        } catch (SQLException e) {
            logger.error("Query failed on shard {}: {}", shardId, sql, e);
            throw new ShardingClientException("Query failed: " + e.getMessage(), e);
        }
    }
    
    /**
     * Execute a query on ALL shards (scatter-gather).
     * Results are merged into a single list.
     * 
     * WARNING: Use sparingly - this hits all shards!
     * 
     * @param sql The SQL query
     * @param params Query parameters
     * @return Combined results from all shards
     */
    public List<Map<String, Object>> queryAllShards(String sql, Object... params) 
            throws ShardingClientException {
        
        List<Map<String, Object>> allResults = new ArrayList<>();
        
        for (Shard shard : shardMap.get().getActiveShards()) {
            try {
                List<Map<String, Object>> shardResults = queryOnShard(shard.getId(), sql, params);
                allResults.addAll(shardResults);
            } catch (ShardingClientException e) {
                logger.warn("Query failed on shard {}, continuing with others", shard.getId(), e);
            }
        }
        
        return allResults;
    }
    
    /**
     * Get the shard ID for a given key.
     * This is a LOCAL operation - no network call.
     */
    public String getShardIdForKey(String shardKey) {
        return shardMap.get().getShardIdForKey(shardKey);
    }
    
    /**
     * Get the current shard map.
     */
    public ShardMap getShardMap() {
        return shardMap.get();
    }
    
    /**
     * Get connection pool statistics for monitoring.
     */
    public Map<String, ShardConnectionManager.PoolStats> getPoolStats() {
        return connectionManager.getPoolStats();
    }
    
    /**
     * Force refresh the shard map from Manager.
     */
    public void refreshShardMap() throws ShardingClientException {
        logger.info("Refreshing shard map from Manager...");
        
        try {
            // Fetch shards for this client app from Manager
            String url = managerUrl + "/api/v1/shards?client_app_id=" + clientAppId;
            HttpGet request = new HttpGet(url);
            request.setHeader("X-Client-App-ID", clientAppId);
            
            try (CloseableHttpResponse response = httpClient.execute(request)) {
                int statusCode = response.getCode();
                String responseBody = EntityUtils.toString(response.getEntity());
                
                if (statusCode != 200) {
                    throw new ShardingClientException(
                        "Failed to fetch shard map: HTTP " + statusCode + " - " + responseBody);
                }
                
                // Parse shard list
                List<Shard> shards = objectMapper.readValue(
                    responseBody, 
                    new TypeReference<List<Shard>>() {}
                );
                
                // Filter shards for this client app
                List<Shard> appShards = new ArrayList<>();
                for (Shard shard : shards) {
                    if (clientAppId.equals(shard.getClientAppId()) || shard.getClientAppId() == null) {
                        appShards.add(shard);
                    }
                }
                
                if (appShards.isEmpty()) {
                    logger.warn("No shards found for client app: {}", clientAppId);
                }
                
                // Create new shard map
                ShardMap newMap = new ShardMap(clientAppId, appShards);
                
                // Initialize connection pools for new shards
                for (Shard shard : appShards) {
                    if (shard.isActive() && !connectionManager.hasPool(shard.getId())) {
                        connectionManager.initializePool(shard);
                    }
                }
                
                // Update cached shard map atomically
                ShardMap oldMap = shardMap.getAndSet(newMap);
                
                logger.info("Shard map refreshed: {} active shards", newMap.getActiveShards().size());
                
                // Close pools for removed shards
                if (oldMap != null) {
                    for (Shard oldShard : oldMap.getAllShards()) {
                        if (newMap.getShardById(oldShard.getId()) == null) {
                            connectionManager.removePool(oldShard.getId());
                        }
                    }
                }
            }
            
        } catch (IOException e) {
            throw new ShardingClientException("Failed to refresh shard map: " + e.getMessage(), e);
        }
    }
    
    private void refreshShardMapSilent() {
        try {
            refreshShardMap();
        } catch (ShardingClientException e) {
            logger.warn("Background shard map refresh failed", e);
        }
    }
    
    private List<Map<String, Object>> mapResultSet(ResultSet rs) throws SQLException {
        List<Map<String, Object>> results = new ArrayList<>();
        ResultSetMetaData metaData = rs.getMetaData();
        int columnCount = metaData.getColumnCount();
        
        while (rs.next()) {
            Map<String, Object> row = new HashMap<>();
            for (int i = 1; i <= columnCount; i++) {
                String columnName = metaData.getColumnLabel(i);
                Object value = rs.getObject(i);
                row.put(columnName, value);
            }
            results.add(row);
        }
        
        return results;
    }
    
    @Override
    public void close() {
        logger.info("Closing ShardingClient...");
        
        // Stop refresh thread
        refreshExecutor.shutdown();
        try {
            if (!refreshExecutor.awaitTermination(5, TimeUnit.SECONDS)) {
                refreshExecutor.shutdownNow();
            }
        } catch (InterruptedException e) {
            refreshExecutor.shutdownNow();
        }
        
        // Close connection pools
        connectionManager.close();
        
        // Close HTTP client
        try {
            httpClient.close();
        } catch (IOException e) {
            logger.warn("Error closing HTTP client", e);
        }
        
        logger.info("ShardingClient closed");
    }
    
    /**
     * Builder for ShardingClient.
     */
    public static class Builder {
        private String managerUrl = "http://localhost:8081";
        private String clientAppId;
        private long refreshIntervalSeconds = 60;
        private int maxPoolSize = 10;
        private int minIdle = 2;
        private long connectionTimeout = 30000;
        private long idleTimeout = 600000;
        private long maxLifetime = 1800000;
        
        /**
         * Set the Manager API URL.
         * The client fetches shard configuration from here on startup.
         */
        public Builder managerUrl(String url) {
            this.managerUrl = url;
            return this;
        }
        
        /**
         * Set the client application ID.
         * This identifies your application to the sharding system.
         */
        public Builder clientAppId(String id) {
            this.clientAppId = id;
            return this;
        }
        
        /**
         * Set the shard map refresh interval in seconds.
         * Default: 60 seconds.
         */
        public Builder refreshInterval(long seconds) {
            this.refreshIntervalSeconds = seconds;
            return this;
        }
        
        /**
         * Set the maximum pool size per shard.
         * Default: 10.
         */
        public Builder maxPoolSize(int size) {
            this.maxPoolSize = size;
            return this;
        }
        
        /**
         * Set the minimum idle connections per shard.
         * Default: 2.
         */
        public Builder minIdle(int idle) {
            this.minIdle = idle;
            return this;
        }
        
        /**
         * Set the connection timeout in milliseconds.
         * Default: 30000 (30 seconds).
         */
        public Builder connectionTimeout(long timeout) {
            this.connectionTimeout = timeout;
            return this;
        }
        
        /**
         * Build the ShardingClient.
         * This will:
         * 1. Fetch shard configuration from Manager
         * 2. Initialize connection pools to all shards
         * 3. Start background refresh thread
         */
        public ShardingClient build() {
            if (clientAppId == null || clientAppId.isEmpty()) {
                throw new IllegalArgumentException("clientAppId is required");
            }
            return new ShardingClient(this);
        }
    }
    
    // ============ Legacy API for backward compatibility ============
    
    /**
     * @deprecated Use {@link Builder} instead
     */
    @Deprecated
    public ShardingClient(String routerUrl) {
        this(routerUrl, "default");
    }
    
    /**
     * @deprecated Use {@link Builder} instead
     */
    @Deprecated
    public ShardingClient(String managerUrl, String clientAppId) {
        this.managerUrl = managerUrl.endsWith("/") 
            ? managerUrl.substring(0, managerUrl.length() - 1) 
            : managerUrl;
        this.clientAppId = clientAppId;
        this.refreshIntervalSeconds = 60;
        
        this.shardMap = new AtomicReference<>();
        this.connectionManager = new ShardConnectionManager();
        this.httpClient = HttpClients.createDefault();
        this.objectMapper = new ObjectMapper();
        this.refreshExecutor = Executors.newSingleThreadScheduledExecutor();
        
        // Try to initialize, but don't fail for legacy usage
        try {
            refreshShardMap();
        } catch (Exception e) {
            logger.warn("Could not initialize shard map (legacy mode)", e);
            // Create empty shard map for legacy router-based usage
            shardMap.set(new ShardMap(clientAppId, new ArrayList<>()));
        }
    }
}
