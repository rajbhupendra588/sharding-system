package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/database"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/models"
	"github.com/sharding-system/pkg/scanner"
	"go.uber.org/zap"
)

// DatabaseHandler handles simplified database management API endpoints
type DatabaseHandler struct {
	dbService           *database.DatabaseService
	manager             *manager.Manager
	logger              *zap.Logger
	databases           map[string]*database.SimpleDatabase
	clusterManager      *scanner.ClusterManager
	multiClusterScanner *scanner.MultiClusterScanner
	scanResults         map[string]models.ScannedDatabase // Store scan results by database ID
	scanResultsMu       sync.RWMutex
}

// NewDatabaseHandler creates a new database handler
func NewDatabaseHandler(
	dbService *database.DatabaseService,
	clusterManager *scanner.ClusterManager,
	multiClusterScanner *scanner.MultiClusterScanner,
	logger *zap.Logger,
) *DatabaseHandler {
	return &DatabaseHandler{
		dbService:           dbService,
		manager:             nil, // Will be set via SetManager
		logger:              logger,
		databases:           make(map[string]*database.SimpleDatabase),
		clusterManager:      clusterManager,
		multiClusterScanner: multiClusterScanner,
		scanResults:         make(map[string]models.ScannedDatabase),
	}
}

// SetManager sets the manager for accessing client apps
func (h *DatabaseHandler) SetManager(mgr *manager.Manager) {
	h.manager = mgr
}

// CreateDatabase handles simplified database creation
// @Summary Create a new sharded database
// @Description Creates a new sharded database with minimal configuration. Uses templates for quick setup.
// @Tags databases
// @Accept json
// @Produce json
// @Param request body database.CreateDatabaseRequest true "Database Configuration"
// @Success 201 {object} database.Database "Database created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/databases [post]
func (h *DatabaseHandler) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	var req database.SimpleCreateDatabaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate name
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Template == "" {
		req.Template = "starter"
	}

	// Create database
	db, err := h.dbService.CreateDatabase(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create database", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store database (in production, use persistent storage)
	h.databases[db.ID] = db

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(db)
}

// convertScannedToSimple converts a ScannedDatabase to SimpleDatabase format
func (h *DatabaseHandler) convertScannedToSimple(scannedDB *models.ScannedDatabase) *database.SimpleDatabase {
	simpleDB := &database.SimpleDatabase{
		ID:               scannedDB.ID,
		Name:             scannedDB.DatabaseName,
		Status:           scannedDB.Status,
		ConnectionString: h.buildConnectionString(scannedDB),
		CreatedAt:        scannedDB.DiscoveredAt,
		UpdatedAt:        scannedDB.DiscoveredAt,
		ShardIDs:         []string{}, // Discovered databases don't have shards yet
	}
	// Add metadata
	if scannedDB.ClusterName != "" {
		simpleDB.Metadata = map[string]interface{}{
			"cluster_id":    scannedDB.ClusterID,
			"cluster_name":  scannedDB.ClusterName,
			"namespace":     scannedDB.Namespace,
			"app_name":      scannedDB.AppName,
			"database_type": scannedDB.DatabaseType,
			"host":          scannedDB.Host,
			"port":          scannedDB.Port,
			"discovered":    true,
		}
	}
	return simpleDB
}

// GetDatabase handles database retrieval
// @Summary Get database by ID
// @Description Retrieves database information by ID (checks manually created, discovered databases, and client apps)
// @Tags databases
// @Accept json
// @Produce json
// @Param id path string true "Database ID"
// @Success 200 {object} database.Database "Database information"
// @Failure 404 {object} map[string]interface{} "Database not found"
// @Router /api/v1/databases/{id} [get]
func (h *DatabaseHandler) GetDatabase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbID := vars["id"]

	// First check manually created databases
	db, ok := h.databases[dbID]
	if !ok {
		// Check discovered databases from scan results
		h.scanResultsMu.RLock()
		scannedDB, found := h.scanResults[dbID]
		h.scanResultsMu.RUnlock()

		if !found {
			// Check if it's a client app database ID (format: "client-app-{appID}")
			if len(dbID) > 11 && dbID[:11] == "client-app-" {
				appID := dbID[11:]
				if h.manager != nil {
					clientAppMgr := h.manager.GetClientAppManager()
					app, err := clientAppMgr.GetClientApp(appID)
					if err == nil && app.DatabaseName != "" && app.DatabaseHost != "" {
						// Build database from client app
						port := 5432
						if app.DatabasePort != "" {
							if p, err := strconv.Atoi(app.DatabasePort); err == nil {
								port = p
							}
						}
						db = &database.SimpleDatabase{
							ID:               dbID,
							Name:             app.DatabaseName,
							DisplayName:      app.Name,
							Description:      app.Description,
							ClientAppID:      app.ID,
							ShardIDs:         app.ShardIDs,
							Status:           "ready",
							ConnectionString: h.buildConnectionStringFromClientApp(app),
							CreatedAt:        app.CreatedAt,
							UpdatedAt:        app.UpdatedAt,
							Metadata: map[string]interface{}{
								"client_app_id":   app.ID,
								"client_app_name": app.Name,
								"namespace":       app.Namespace,
								"cluster_name":    app.ClusterName,
								"from_client_app": true,
								"host":            app.DatabaseHost,
								"port":            port,
								"user":            app.DatabaseUser,
							},
						}
					} else {
						http.Error(w, "database not found", http.StatusNotFound)
						return
					}
				} else {
					http.Error(w, "database not found", http.StatusNotFound)
					return
				}
			} else {
				http.Error(w, "database not found", http.StatusNotFound)
				return
			}
		} else {
			// Convert discovered database to SimpleDatabase format
			db = h.convertScannedToSimple(&scannedDB)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(db)
}

// ListDatabases handles database listing
// @Summary List all databases
// @Description Returns a list of all databases (manually created, discovered from clusters, and from registered client apps)
// @Tags databases
// @Accept json
// @Produce json
// @Success 200 {array} database.Database "List of databases"
// @Router /api/v1/databases [get]
func (h *DatabaseHandler) ListDatabases(w http.ResponseWriter, r *http.Request) {
	// Track seen database IDs to avoid duplicates
	seenDBs := make(map[string]bool)
	databases := make([]*database.SimpleDatabase, 0)

	// Start with manually created databases
	for _, db := range h.databases {
		databases = append(databases, db)
		seenDBs[db.ID] = true
	}

	// Add discovered databases from clusters
	h.scanResultsMu.RLock()
	for _, scannedDB := range h.scanResults {
		if !seenDBs[scannedDB.ID] {
			databases = append(databases, h.convertScannedToSimple(&scannedDB))
			seenDBs[scannedDB.ID] = true
		}
	}
	h.scanResultsMu.RUnlock()

	// Add databases from registered client apps that have database information
	if h.manager != nil {
		clientAppMgr := h.manager.GetClientAppManager()
		clientApps, err := clientAppMgr.ListClientApps()
		if err == nil {
			for _, app := range clientApps {
				// Only include apps that have database information
				if app.DatabaseName != "" && app.DatabaseHost != "" {
					// Generate a consistent ID for this database from client app
					dbID := fmt.Sprintf("client-app-%s", app.ID)
					if !seenDBs[dbID] {
						// Build connection string
						connStr := h.buildConnectionStringFromClientApp(app)

						// Parse port
						port := 5432 // Default PostgreSQL port
						if app.DatabasePort != "" {
							if p, err := strconv.Atoi(app.DatabasePort); err == nil {
								port = p
							}
						}

						db := &database.SimpleDatabase{
							ID:               dbID,
							Name:             app.DatabaseName,
							DisplayName:      app.Name,
							Description:      app.Description,
							ClientAppID:      app.ID,
							ShardIDs:         app.ShardIDs,
							Status:           "ready", // Client apps with DB info are considered ready
							ConnectionString: connStr,
							CreatedAt:        app.CreatedAt,
							UpdatedAt:        app.UpdatedAt,
							Metadata: map[string]interface{}{
								"client_app_id":   app.ID,
								"client_app_name": app.Name,
								"namespace":       app.Namespace,
								"cluster_name":    app.ClusterName,
								"from_client_app": true,
								"host":            app.DatabaseHost,
								"port":            port,
								"user":            app.DatabaseUser,
							},
						}
						databases = append(databases, db)
						seenDBs[dbID] = true
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(databases)
}

// buildConnectionStringFromClientApp builds a connection string from client app database info
func (h *DatabaseHandler) buildConnectionStringFromClientApp(app *manager.ClientAppInfo) string {
	// Build from components
	port := app.DatabasePort
	if port == "" {
		port = "5432" // Default PostgreSQL port
	}

	// Try to determine database type from port or default to postgresql
	dbType := "postgresql"
	if port == "3306" {
		dbType = "mysql"
	}

	user := app.DatabaseUser
	if user == "" {
		user = "postgres" // Default user
	}

	host := app.DatabaseHost
	if host == "" {
		host = "localhost"
	}

	if dbType == "postgresql" {
		return fmt.Sprintf("postgres://%s@%s:%s/%s", user, host, port, app.DatabaseName)
	}
	return fmt.Sprintf("%s://%s@%s:%s/%s", dbType, user, host, port, app.DatabaseName)
}

// buildConnectionString builds a connection string from scanned database info
func (h *DatabaseHandler) buildConnectionString(db *models.ScannedDatabase) string {
	if db.DatabaseType == "postgresql" {
		return fmt.Sprintf("postgres://%s@%s:%d/%s", db.Username, db.Host, db.Port, db.Database)
	}
	return fmt.Sprintf("%s://%s@%s:%d/%s", db.DatabaseType, db.Username, db.Host, db.Port, db.Database)
}

// UpdateScanResults updates the stored scan results
func (h *DatabaseHandler) UpdateScanResults(results []models.ScannedDatabase) {
	h.scanResultsMu.Lock()
	defer h.scanResultsMu.Unlock()

	for _, db := range results {
		h.scanResults[db.ID] = db
	}
	h.logger.Info("updated scan results", zap.Int("count", len(results)))
}

// ListTemplates handles template listing
// @Summary List available database templates
// @Description Returns all available database templates (starter, production, enterprise)
// @Tags databases
// @Accept json
// @Produce json
// @Success 200 {array} database.DatabaseTemplate "List of templates"
// @Router /api/v1/databases/templates [get]
func (h *DatabaseHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates := database.ListTemplates()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// GetDatabaseStatus handles database status retrieval
// @Summary Get database status
// @Description Returns the current status of a database including shard information (checks both manually created and discovered databases)
// @Tags databases
// @Accept json
// @Produce json
// @Param id path string true "Database ID"
// @Success 200 {object} map[string]interface{} "Database status"
// @Failure 404 {object} map[string]interface{} "Database not found"
// @Router /api/v1/databases/{id}/status [get]
func (h *DatabaseHandler) GetDatabaseStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbID := vars["id"]

	// First check manually created databases
	db, ok := h.databases[dbID]
	if !ok {
		// Check discovered databases from scan results
		h.scanResultsMu.RLock()
		scannedDB, found := h.scanResults[dbID]
		h.scanResultsMu.RUnlock()

		if !found {
			http.Error(w, "database not found", http.StatusNotFound)
			return
		}

		// Convert discovered database to SimpleDatabase format
		db = h.convertScannedToSimple(&scannedDB)
	}

	status := map[string]interface{}{
		"id":                db.ID,
		"name":              db.Name,
		"status":            db.Status,
		"shard_count":       len(db.ShardIDs),
		"connection_string": db.ConnectionString,
		"created_at":        db.CreatedAt,
		"updated_at":        db.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// DatabaseStats represents aggregated database statistics
type DatabaseStats struct {
	TotalDatabases int            `json:"total_databases"`
	ByStatus       map[string]int `json:"by_status"`
	ByType         map[string]int `json:"by_type"`
}

// GetDatabaseStats handles database statistics retrieval
// @Summary Get database statistics
// @Description Returns aggregated statistics about databases
// @Tags databases
// @Accept json
// @Produce json
// @Success 200 {object} DatabaseStats "Database statistics"
// @Router /api/v1/databases/stats [get]
func (h *DatabaseHandler) GetDatabaseStats(w http.ResponseWriter, r *http.Request) {
	stats := DatabaseStats{
		ByStatus: make(map[string]int),
		ByType:   make(map[string]int),
	}

	// Count manually created databases
	for _, db := range h.databases {
		stats.TotalDatabases++
		stats.ByStatus[db.Status]++
		// Manually created databases are usually postgresql by default in this system
		// or we could check metadata if available. For now, we'll assume postgresql
		// or check if there's a type field. SimpleDatabase doesn't have a Type field directly
		// but we can infer or leave it generic.
		// Let's check if we can get type from metadata
		if dbType, ok := db.Metadata["database_type"].(string); ok {
			stats.ByType[dbType]++
		} else {
			stats.ByType["postgresql"]++ // Default assumption for manual creation in this context
		}
	}

	// Count discovered databases
	h.scanResultsMu.RLock()
	for _, db := range h.scanResults {
		// Avoid double counting if ID exists in both (though they shouldn't usually)
		if _, exists := h.databases[db.ID]; !exists {
			stats.TotalDatabases++
			stats.ByStatus[db.Status]++
			stats.ByType[db.DatabaseType]++
		}
	}
	h.scanResultsMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// SetupDatabaseRoutes sets up database management routes
func SetupDatabaseRoutes(router *mux.Router, handler *DatabaseHandler) {
	router.HandleFunc("/api/v1/databases", handler.CreateDatabase).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/databases", handler.ListDatabases).Methods("GET", "OPTIONS")
	// Register specific routes before parameterized routes to avoid conflicts
	router.HandleFunc("/api/v1/databases/templates", handler.ListTemplates).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/databases/stats", handler.GetDatabaseStats).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/databases/{id}/status", handler.GetDatabaseStatus).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/databases/{id}", handler.GetDatabase).Methods("GET", "OPTIONS")
}
