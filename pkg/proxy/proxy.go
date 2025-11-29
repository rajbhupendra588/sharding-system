package proxy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/sharding-system/pkg/hashing"
	"github.com/sharding-system/pkg/models"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// ShardingProxy is the main proxy server that intercepts database connections
// and routes queries to the appropriate shards automatically.
//
// This enables ZERO-CODE sharding - applications just change their connection
// string to point to this proxy instead of the database directly.
type ShardingProxy struct {
	config       *ProxyConfig
	logger       *zap.Logger
	sqlParser    *SQLParser
	hashFunc     hashing.HashFunction
	
	// Shard connections - pooled connections to each shard
	shardPools   map[string]*sql.DB
	shardPoolsMu sync.RWMutex
	
	// Shard metadata from manager
	shards       []models.Shard
	shardsMu     sync.RWMutex
	
	// Listeners
	dbListener   net.Listener
	adminServer  *http.Server
	
	// Lifecycle
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewShardingProxy creates a new sharding proxy
func NewShardingProxy(config *ProxyConfig, logger *zap.Logger) *ShardingProxy {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ShardingProxy{
		config:     config,
		logger:     logger,
		sqlParser:  NewSQLParser(),
		hashFunc:   hashing.NewHashFunction("murmur3"),
		shardPools: make(map[string]*sql.DB),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start starts the proxy server
func (p *ShardingProxy) Start() error {
	p.logger.Info("starting sharding proxy",
		zap.String("db_listen", p.config.ListenAddr),
		zap.String("admin_listen", p.config.AdminAddr))
	
	// Load shard configuration from manager
	if err := p.refreshShards(); err != nil {
		p.logger.Warn("failed to load shards from manager, will retry", zap.Error(err))
	}
	
	// Start background shard refresh
	p.wg.Add(1)
	go p.shardRefreshLoop()
	
	// Start admin HTTP server
	if err := p.startAdminServer(); err != nil {
		return fmt.Errorf("failed to start admin server: %w", err)
	}
	
	// Start database proxy listener
	listener, err := net.Listen("tcp", p.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", p.config.ListenAddr, err)
	}
	p.dbListener = listener
	
	p.logger.Info("sharding proxy started",
		zap.String("db_addr", p.config.ListenAddr),
		zap.String("admin_addr", p.config.AdminAddr))
	
	// Accept connections
	p.wg.Add(1)
	go p.acceptLoop()
	
	return nil
}

// Stop stops the proxy server
func (p *ShardingProxy) Stop() error {
	p.logger.Info("stopping sharding proxy")
	
	p.cancel()
	
	if p.dbListener != nil {
		p.dbListener.Close()
	}
	
	if p.adminServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		p.adminServer.Shutdown(ctx)
	}
	
	// Close shard pools
	p.shardPoolsMu.Lock()
	for _, pool := range p.shardPools {
		pool.Close()
	}
	p.shardPoolsMu.Unlock()
	
	p.wg.Wait()
	p.logger.Info("sharding proxy stopped")
	
	return nil
}

// acceptLoop accepts incoming connections
func (p *ShardingProxy) acceptLoop() {
	defer p.wg.Done()
	
	for {
		conn, err := p.dbListener.Accept()
		if err != nil {
			select {
			case <-p.ctx.Done():
				return
			default:
				p.logger.Error("failed to accept connection", zap.Error(err))
				continue
			}
		}
		
		p.wg.Add(1)
		go p.handleConnection(conn)
	}
}

// handleConnection handles a single client connection
func (p *ShardingProxy) handleConnection(conn net.Conn) {
	defer p.wg.Done()
	defer conn.Close()
	
	clientAddr := conn.RemoteAddr().String()
	p.logger.Debug("new connection", zap.String("client", clientAddr))
	
	// For now, use a simple line-based protocol for demonstration
	// In production, this would implement the full PostgreSQL wire protocol
	// using a library like jackc/pgproto3
	
	// Read the query
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		if err != io.EOF {
			p.logger.Error("failed to read from connection", zap.Error(err))
		}
		return
	}
	
	query := string(buf[:n])
	p.logger.Debug("received query", zap.String("query", query))
	
	// Execute the query
	result, err := p.ExecuteQuery(context.Background(), "default_db", query)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("ERROR: %s\n", err.Error())))
		return
	}
	
	// Return result
	resultJSON, _ := json.Marshal(result)
	conn.Write(resultJSON)
}

// ExecuteQuery executes a query with automatic shard routing
func (p *ShardingProxy) ExecuteQuery(ctx context.Context, database string, sql string) (*QueryResult, error) {
	startTime := time.Now()
	
	// Get app config
	appConfig := p.config.GetAppConfig(database)
	if appConfig == nil {
		// No sharding rules, route to default
		return p.executeOnAllShards(ctx, sql)
	}
	
	// Extract table from query
	table := ExtractTableFromSQL(sql)
	if table == "" {
		// Can't determine table, broadcast to all shards
		return p.executeOnAllShards(ctx, sql)
	}
	
	// Get sharding rule for this table
	rule := appConfig.GetShardingRule(table)
	if rule == nil {
		// No sharding rule for this table, broadcast
		return p.executeOnAllShards(ctx, sql)
	}
	
	// Handle broadcast strategy
	if rule.Strategy == "broadcast" {
		return p.executeOnAllShards(ctx, sql)
	}
	
	// Parse query to extract shard key
	parsed, err := p.sqlParser.Parse(sql, rule.ShardKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}
	
	p.logger.Debug("parsed query",
		zap.String("table", table),
		zap.String("shard_key", rule.ShardKey),
		zap.String("shard_value", parsed.ShardValue),
		zap.Bool("can_route", parsed.CanRoute))
	
	// If we can route to a specific shard
	if parsed.CanRoute && parsed.ShardValue != "" {
		shard := p.getShardForKey(parsed.ShardValue)
		if shard == nil {
			return nil, fmt.Errorf("no shard found for key: %s", parsed.ShardValue)
		}
		
		result, err := p.executeOnShard(ctx, shard, sql)
		if err != nil {
			return nil, err
		}
		
		result.RoutedTo = shard.ID
		result.LatencyMs = float64(time.Since(startTime).Milliseconds())
		return result, nil
	}
	
	// Cross-shard query - scatter-gather
	return p.executeOnAllShards(ctx, sql)
}

// getShardForKey returns the shard that owns a given key
func (p *ShardingProxy) getShardForKey(key string) *models.Shard {
	p.shardsMu.RLock()
	defer p.shardsMu.RUnlock()
	
	if len(p.shards) == 0 {
		return nil
	}
	
	// Hash the key
	hash := p.hashFunc.Hash(key)
	
	// Find the shard that owns this hash
	for i := range p.shards {
		shard := &p.shards[i]
		if shard.Status != "active" {
			continue
		}
		
		// Check if hash falls in this shard's range
		if hash >= shard.HashRangeStart && hash <= shard.HashRangeEnd {
			return shard
		}
	}
	
	// Fallback to first active shard
	for i := range p.shards {
		if p.shards[i].Status == "active" {
			return &p.shards[i]
		}
	}
	
	return nil
}

// executeOnShard executes a query on a specific shard
func (p *ShardingProxy) executeOnShard(ctx context.Context, shard *models.Shard, sql string) (*QueryResult, error) {
	pool := p.getOrCreatePool(shard)
	if pool == nil {
		return nil, fmt.Errorf("no connection pool for shard: %s", shard.ID)
	}
	
	rows, err := pool.QueryContext(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query failed on shard %s: %w", shard.ID, err)
	}
	defer rows.Close()
	
	return p.scanResults(rows)
}

// executeOnAllShards executes a query on all shards (scatter-gather)
func (p *ShardingProxy) executeOnAllShards(ctx context.Context, sql string) (*QueryResult, error) {
	p.shardsMu.RLock()
	shards := make([]models.Shard, len(p.shards))
	copy(shards, p.shards)
	p.shardsMu.RUnlock()
	
	if len(shards) == 0 {
		return nil, fmt.Errorf("no shards available")
	}
	
	// Execute on all shards in parallel
	type shardResult struct {
		shardID string
		result  *QueryResult
		err     error
	}
	
	results := make(chan shardResult, len(shards))
	
	for i := range shards {
		shard := &shards[i]
		if shard.Status != "active" {
			continue
		}
		
		go func(s *models.Shard) {
			result, err := p.executeOnShard(ctx, s, sql)
			results <- shardResult{shardID: s.ID, result: result, err: err}
		}(shard)
	}
	
	// Collect results
	combined := &QueryResult{
		Rows:     make([]map[string]interface{}, 0),
		RoutedTo: "all_shards",
	}
	
	activeShards := 0
	for i := range shards {
		if shards[i].Status == "active" {
			activeShards++
		}
	}
	
	for i := 0; i < activeShards; i++ {
		select {
		case sr := <-results:
			if sr.err != nil {
				p.logger.Warn("query failed on shard", 
					zap.String("shard", sr.shardID),
					zap.Error(sr.err))
				continue
			}
			combined.Rows = append(combined.Rows, sr.result.Rows...)
			combined.RowCount += sr.result.RowCount
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	
	return combined, nil
}

// scanResults scans query results into a QueryResult
func (p *ShardingProxy) scanResults(rows *sql.Rows) (*QueryResult, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	
	result := &QueryResult{
		Columns: columns,
		Rows:    make([]map[string]interface{}, 0),
	}
	
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		result.Rows = append(result.Rows, row)
		result.RowCount++
	}
	
	return result, rows.Err()
}

// getOrCreatePool gets or creates a connection pool for a shard
func (p *ShardingProxy) getOrCreatePool(shard *models.Shard) *sql.DB {
	p.shardPoolsMu.RLock()
	pool, exists := p.shardPools[shard.ID]
	p.shardPoolsMu.RUnlock()
	
	if exists {
		return pool
	}
	
	p.shardPoolsMu.Lock()
	defer p.shardPoolsMu.Unlock()
	
	// Double-check after acquiring write lock
	if pool, exists = p.shardPools[shard.ID]; exists {
		return pool
	}
	
	// Create new pool
	db, err := sql.Open("postgres", shard.PrimaryEndpoint)
	if err != nil {
		p.logger.Error("failed to create connection pool",
			zap.String("shard", shard.ID),
			zap.Error(err))
		return nil
	}
	
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	
	p.shardPools[shard.ID] = db
	p.logger.Info("created connection pool for shard", zap.String("shard", shard.ID))
	
	return db
}

// refreshShards loads shard configuration from the manager
func (p *ShardingProxy) refreshShards() error {
	url := p.config.ManagerURL + "/api/v1/shards"
	
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch shards: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("manager returned status %d", resp.StatusCode)
	}
	
	var shards []models.Shard
	if err := json.NewDecoder(resp.Body).Decode(&shards); err != nil {
		return fmt.Errorf("failed to decode shards: %w", err)
	}
	
	p.shardsMu.Lock()
	p.shards = shards
	p.shardsMu.Unlock()
	
	p.logger.Info("refreshed shard configuration", zap.Int("shard_count", len(shards)))
	
	return nil
}

// shardRefreshLoop periodically refreshes shard configuration
func (p *ShardingProxy) shardRefreshLoop() {
	defer p.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := p.refreshShards(); err != nil {
				p.logger.Warn("failed to refresh shards", zap.Error(err))
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// QueryResult represents the result of a query
type QueryResult struct {
	Columns   []string                 `json:"columns,omitempty"`
	Rows      []map[string]interface{} `json:"rows"`
	RowCount  int                      `json:"row_count"`
	RoutedTo  string                   `json:"routed_to"` // Shard ID or "all_shards"
	LatencyMs float64                  `json:"latency_ms"`
}

