package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// startAdminServer starts the admin HTTP server for managing sharding rules
func (p *ShardingProxy) startAdminServer() error {
	router := mux.NewRouter()
	
	// CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	
	// Health check
	router.HandleFunc("/health", p.healthHandler).Methods("GET")
	
	// Sharding rules management
	router.HandleFunc("/api/v1/rules", p.listRulesHandler).Methods("GET")
	router.HandleFunc("/api/v1/rules/{database}", p.getRulesHandler).Methods("GET")
	router.HandleFunc("/api/v1/rules/{database}", p.createRulesHandler).Methods("POST")
	router.HandleFunc("/api/v1/rules/{database}/{table}", p.updateRuleHandler).Methods("PUT")
	router.HandleFunc("/api/v1/rules/{database}/{table}", p.deleteRuleHandler).Methods("DELETE")
	
	// Query testing endpoint
	router.HandleFunc("/api/v1/query", p.testQueryHandler).Methods("POST")
	
	// Stats
	router.HandleFunc("/api/v1/stats", p.statsHandler).Methods("GET")
	
	p.adminServer = &http.Server{
		Addr:    p.config.AdminAddr,
		Handler: router,
	}
	
	go func() {
		if err := p.adminServer.ListenAndServe(); err != http.ErrServerClosed {
			p.logger.Error("admin server error", zap.Error(err))
		}
	}()
	
	return nil
}

func (p *ShardingProxy) healthHandler(w http.ResponseWriter, r *http.Request) {
	p.shardsMu.RLock()
	shardCount := len(p.shards)
	p.shardsMu.RUnlock()
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "healthy",
		"shard_count": shardCount,
	})
}

// listRulesHandler returns all sharding rules for all databases
func (p *ShardingProxy) listRulesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p.config.ClientApps)
}

// getRulesHandler returns sharding rules for a specific database
func (p *ShardingProxy) getRulesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	
	config := p.config.GetAppConfig(database)
	if config == nil {
		http.Error(w, "database not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// CreateRulesRequest represents a request to create sharding rules
type CreateRulesRequest struct {
	Name          string         `json:"name"`
	ShardingRules []ShardingRule `json:"sharding_rules"`
}

// createRulesHandler creates sharding rules for a database
func (p *ShardingProxy) createRulesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	
	var req CreateRulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	config := &ClientAppConfig{
		ID:            database,
		Name:          req.Name,
		Database:      database,
		ShardingRules: req.ShardingRules,
	}
	
	p.config.SetAppConfig(database, config)
	
	p.logger.Info("created sharding rules",
		zap.String("database", database),
		zap.Int("rule_count", len(req.ShardingRules)))
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(config)
}

// updateRuleHandler updates a single sharding rule
func (p *ShardingProxy) updateRuleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	table := vars["table"]
	
	config := p.config.GetAppConfig(database)
	if config == nil {
		http.Error(w, "database not found", http.StatusNotFound)
		return
	}
	
	var rule ShardingRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	rule.Table = table
	config.AddShardingRule(rule)
	
	p.logger.Info("updated sharding rule",
		zap.String("database", database),
		zap.String("table", table),
		zap.String("shard_key", rule.ShardKey))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

// deleteRuleHandler deletes a sharding rule
func (p *ShardingProxy) deleteRuleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	database := vars["database"]
	table := vars["table"]
	
	config := p.config.GetAppConfig(database)
	if config == nil {
		http.Error(w, "database not found", http.StatusNotFound)
		return
	}
	
	config.RemoveShardingRule(table)
	
	p.logger.Info("deleted sharding rule",
		zap.String("database", database),
		zap.String("table", table))
	
	w.WriteHeader(http.StatusNoContent)
}

// TestQueryRequest represents a request to test query routing
type TestQueryRequest struct {
	Database string `json:"database"`
	Query    string `json:"query"`
}

// testQueryHandler tests how a query would be routed
func (p *ShardingProxy) testQueryHandler(w http.ResponseWriter, r *http.Request) {
	var req TestQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Get app config
	appConfig := p.config.GetAppConfig(req.Database)
	
	// Extract table
	table := ExtractTableFromSQL(req.Query)
	
	result := map[string]interface{}{
		"query":     req.Query,
		"database":  req.Database,
		"table":     table,
		"routing":   "unknown",
		"shard_key": "",
		"shard_value": "",
	}
	
	if appConfig == nil {
		result["routing"] = "broadcast (no sharding rules)"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}
	
	rule := appConfig.GetShardingRule(table)
	if rule == nil {
		result["routing"] = "broadcast (no rule for table)"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}
	
	result["shard_key"] = rule.ShardKey
	result["strategy"] = rule.Strategy
	
	if rule.Strategy == "broadcast" {
		result["routing"] = "broadcast (strategy)"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}
	
	// Parse query
	parsed, err := p.sqlParser.Parse(req.Query, rule.ShardKey)
	if err != nil {
		result["error"] = err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}
	
	result["parsed"] = parsed
	
	if parsed.CanRoute && parsed.ShardValue != "" {
		result["shard_value"] = parsed.ShardValue
		
		shard := p.getShardForKey(parsed.ShardValue)
		if shard != nil {
			result["routing"] = "single_shard"
			result["target_shard"] = shard.ID
			result["target_endpoint"] = shard.PrimaryEndpoint
		} else {
			result["routing"] = "no_shard_found"
		}
	} else {
		result["routing"] = "scatter_gather (shard key not in WHERE clause)"
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// statsHandler returns proxy statistics
func (p *ShardingProxy) statsHandler(w http.ResponseWriter, r *http.Request) {
	p.shardsMu.RLock()
	shards := make([]map[string]interface{}, 0, len(p.shards))
	for _, shard := range p.shards {
		shards = append(shards, map[string]interface{}{
			"id":       shard.ID,
			"name":     shard.Name,
			"status":   shard.Status,
			"endpoint": shard.PrimaryEndpoint,
		})
	}
	p.shardsMu.RUnlock()
	
	p.shardPoolsMu.RLock()
	poolCount := len(p.shardPools)
	p.shardPoolsMu.RUnlock()
	
	stats := map[string]interface{}{
		"shards":           shards,
		"shard_count":      len(shards),
		"connection_pools": poolCount,
		"databases":        len(p.config.ClientApps),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// APIResponse is a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIResponse{Success: false, Error: message})
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: data})
}

// GenerateProxyConnectionString generates the connection string for clients
func GenerateProxyConnectionString(proxyHost string, proxyPort int, database string) string {
	return fmt.Sprintf("postgresql://%s:%d/%s", proxyHost, proxyPort, database)
}

