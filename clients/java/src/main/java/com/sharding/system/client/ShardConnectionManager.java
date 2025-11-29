package com.sharding.system.client;

import com.sharding.system.client.model.Shard;
import com.zaxxer.hikari.HikariConfig;
import com.zaxxer.hikari.HikariDataSource;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.SQLException;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Manages connection pools for all database shards.
 * 
 * Each shard has its own HikariCP connection pool for optimal performance.
 * Connections are made DIRECTLY to the database - no proxy involved.
 */
public class ShardConnectionManager implements AutoCloseable {
    
    private static final Logger logger = LoggerFactory.getLogger(ShardConnectionManager.class);
    
    // Connection pools keyed by shard ID
    private final Map<String, HikariDataSource> connectionPools;
    
    // Connection pool configuration
    private final int maxPoolSize;
    private final int minIdle;
    private final long connectionTimeout;
    private final long idleTimeout;
    private final long maxLifetime;
    
    /**
     * Creates a new connection manager with default pool settings.
     */
    public ShardConnectionManager() {
        this(10, 2, 30000, 600000, 1800000);
    }
    
    /**
     * Creates a new connection manager with custom pool settings.
     * 
     * @param maxPoolSize Maximum number of connections per shard
     * @param minIdle Minimum number of idle connections per shard
     * @param connectionTimeout Connection timeout in milliseconds
     * @param idleTimeout Idle connection timeout in milliseconds
     * @param maxLifetime Maximum connection lifetime in milliseconds
     */
    public ShardConnectionManager(int maxPoolSize, int minIdle, long connectionTimeout, 
                                   long idleTimeout, long maxLifetime) {
        this.connectionPools = new ConcurrentHashMap<>();
        this.maxPoolSize = maxPoolSize;
        this.minIdle = minIdle;
        this.connectionTimeout = connectionTimeout;
        this.idleTimeout = idleTimeout;
        this.maxLifetime = maxLifetime;
    }
    
    /**
     * Initialize connection pool for a shard.
     * 
     * @param shard The shard configuration
     */
    public void initializePool(Shard shard) {
        if (connectionPools.containsKey(shard.getId())) {
            logger.debug("Connection pool already exists for shard: {}", shard.getId());
            return;
        }
        
        logger.info("Initializing connection pool for shard: {} -> {}", shard.getId(), shard.getEndpoint());
        
        HikariConfig config = new HikariConfig();
        config.setJdbcUrl(convertToJdbcUrl(shard.getEndpoint()));
        config.setPoolName("shard-" + shard.getId());
        config.setMaximumPoolSize(maxPoolSize);
        config.setMinimumIdle(minIdle);
        config.setConnectionTimeout(connectionTimeout);
        config.setIdleTimeout(idleTimeout);
        config.setMaxLifetime(maxLifetime);
        
        // PostgreSQL specific optimizations
        config.addDataSourceProperty("cachePrepStmts", "true");
        config.addDataSourceProperty("prepStmtCacheSize", "250");
        config.addDataSourceProperty("prepStmtCacheSqlLimit", "2048");
        config.addDataSourceProperty("useServerPrepStmts", "true");
        
        HikariDataSource dataSource = new HikariDataSource(config);
        connectionPools.put(shard.getId(), dataSource);
        
        logger.info("Connection pool initialized for shard: {}", shard.getId());
    }
    
    /**
     * Get a connection from the pool for a specific shard.
     * 
     * @param shardId The shard ID
     * @return A database connection
     * @throws SQLException if unable to get connection
     */
    public Connection getConnection(String shardId) throws SQLException {
        HikariDataSource pool = connectionPools.get(shardId);
        if (pool == null) {
            throw new SQLException("No connection pool found for shard: " + shardId);
        }
        return pool.getConnection();
    }
    
    /**
     * Get the DataSource for a specific shard.
     * 
     * @param shardId The shard ID
     * @return The DataSource for this shard
     */
    public DataSource getDataSource(String shardId) {
        return connectionPools.get(shardId);
    }
    
    /**
     * Check if a connection pool exists for a shard.
     */
    public boolean hasPool(String shardId) {
        return connectionPools.containsKey(shardId);
    }
    
    /**
     * Remove and close a connection pool for a shard.
     */
    public void removePool(String shardId) {
        HikariDataSource pool = connectionPools.remove(shardId);
        if (pool != null) {
            logger.info("Closing connection pool for shard: {}", shardId);
            pool.close();
        }
    }
    
    /**
     * Convert a PostgreSQL connection string to JDBC URL.
     * Supports formats:
     * - postgresql://user:pass@host:port/database
     * - postgres://user:pass@host:port/database
     * - jdbc:postgresql://host:port/database (passthrough)
     */
    private String convertToJdbcUrl(String endpoint) {
        if (endpoint.startsWith("jdbc:")) {
            return endpoint;
        }
        
        // Convert postgresql:// or postgres:// to jdbc:postgresql://
        if (endpoint.startsWith("postgresql://") || endpoint.startsWith("postgres://")) {
            String url = endpoint.replaceFirst("postgres(ql)?://", "jdbc:postgresql://");
            return url;
        }
        
        // Assume it's a host:port/database format
        return "jdbc:postgresql://" + endpoint;
    }
    
    /**
     * Get pool statistics for monitoring.
     */
    public Map<String, PoolStats> getPoolStats() {
        Map<String, PoolStats> stats = new ConcurrentHashMap<>();
        for (Map.Entry<String, HikariDataSource> entry : connectionPools.entrySet()) {
            HikariDataSource pool = entry.getValue();
            stats.put(entry.getKey(), new PoolStats(
                pool.getHikariPoolMXBean().getActiveConnections(),
                pool.getHikariPoolMXBean().getIdleConnections(),
                pool.getHikariPoolMXBean().getTotalConnections(),
                pool.getHikariPoolMXBean().getThreadsAwaitingConnection()
            ));
        }
        return stats;
    }
    
    @Override
    public void close() {
        logger.info("Closing all connection pools...");
        for (Map.Entry<String, HikariDataSource> entry : connectionPools.entrySet()) {
            try {
                entry.getValue().close();
                logger.debug("Closed connection pool for shard: {}", entry.getKey());
            } catch (Exception e) {
                logger.error("Error closing connection pool for shard: {}", entry.getKey(), e);
            }
        }
        connectionPools.clear();
    }
    
    /**
     * Pool statistics for monitoring.
     */
    public static class PoolStats {
        public final int activeConnections;
        public final int idleConnections;
        public final int totalConnections;
        public final int threadsWaiting;
        
        public PoolStats(int active, int idle, int total, int waiting) {
            this.activeConnections = active;
            this.idleConnections = idle;
            this.totalConnections = total;
            this.threadsWaiting = waiting;
        }
        
        @Override
        public String toString() {
            return "PoolStats{active=" + activeConnections + ", idle=" + idleConnections + 
                   ", total=" + totalConnections + ", waiting=" + threadsWaiting + "}";
        }
    }
}

