package cluster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Cluster represents a Kubernetes cluster (cloud or on-premise)
type Cluster struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"` // "cloud" (aws, gcp, azure) or "onprem"
	Provider    string            `json:"provider,omitempty"` // "aws", "gcp", "azure", "onprem", etc.
	Endpoint    string            `json:"endpoint,omitempty"` // K8s API endpoint
	Kubeconfig  string            `json:"-"`                 // Kubeconfig content (not serialized)
	KubeconfigPath string         `json:"kubeconfig_path,omitempty"` // Path to kubeconfig file
	Status      string            `json:"status"`            // "connected", "disconnected", "error"
	LastSeen    time.Time         `json:"last_seen"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ClusterManager manages multiple Kubernetes clusters
type ClusterManager struct {
	clusters map[string]*Cluster
	clients  map[string]*kubernetes.Clientset // Cache of K8s clients
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewClusterManager creates a new cluster manager
func NewClusterManager(logger *zap.Logger) *ClusterManager {
	return &ClusterManager{
		clusters: make(map[string]*Cluster),
		clients:  make(map[string]*kubernetes.Clientset),
		logger:   logger,
	}
}

// RegisterCluster registers a new Kubernetes cluster
func (cm *ClusterManager) RegisterCluster(ctx context.Context, name, clusterType, provider, kubeconfigPath, kubeconfigContent string, metadata map[string]string) (*Cluster, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if cluster with same name exists
	for _, cluster := range cm.clusters {
		if cluster.Name == name {
			return nil, fmt.Errorf("cluster with name '%s' already exists", name)
		}
	}

	cluster := &Cluster{
		ID:            uuid.New().String(),
		Name:          name,
		Type:          clusterType,
		Provider:      provider,
		Kubeconfig:    kubeconfigContent,
		KubeconfigPath: kubeconfigPath,
		Status:        "disconnected",
		Metadata:      metadata,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Try to connect to validate the cluster
	client, err := cm.createClient(cluster)
	if err != nil {
		cluster.Status = "error"
		cm.logger.Warn("failed to connect to cluster during registration",
			zap.String("cluster_id", cluster.ID),
			zap.String("name", name),
			zap.Error(err))
	} else {
		cluster.Status = "connected"
		cluster.LastSeen = time.Now()
		cm.clients[cluster.ID] = client
	}

	cm.clusters[cluster.ID] = cluster
	cm.logger.Info("registered cluster",
		zap.String("cluster_id", cluster.ID),
		zap.String("name", name),
		zap.String("type", clusterType),
		zap.String("provider", provider))

	return cluster, nil
}

// GetCluster retrieves a cluster by ID
func (cm *ClusterManager) GetCluster(clusterID string) (*Cluster, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cluster, exists := cm.clusters[clusterID]
	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", clusterID)
	}

	return cluster, nil
}

// ListClusters returns all registered clusters
func (cm *ClusterManager) ListClusters() []*Cluster {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clusters := make([]*Cluster, 0, len(cm.clusters))
	for _, cluster := range cm.clusters {
		clusters = append(clusters, cluster)
	}

	return clusters
}

// GetClient returns a Kubernetes client for a cluster
func (cm *ClusterManager) GetClient(clusterID string) (*kubernetes.Clientset, error) {
	cm.mu.RLock()
	client, exists := cm.clients[clusterID]
	cm.mu.RUnlock()

	if exists && client != nil {
		return client, nil
	}

	// Client not cached, create it
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cluster, exists := cm.clusters[clusterID]
	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", clusterID)
	}

	client, err := cm.createClient(cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for cluster %s: %w", clusterID, err)
	}

	cm.clients[clusterID] = client
	cluster.Status = "connected"
	cluster.LastSeen = time.Now()
	cluster.UpdatedAt = time.Now()

	return client, nil
}

// createClient creates a Kubernetes client from cluster config
func (cm *ClusterManager) createClient(cluster *Cluster) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// If kubeconfig content is provided, use it
	if cluster.Kubeconfig != "" {
		config, err = clientcmd.RESTConfigFromKubeConfig([]byte(cluster.Kubeconfig))
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
		}
	} else if cluster.KubeconfigPath != "" {
		// Use kubeconfig file path
		config, err = clientcmd.BuildConfigFromFlags("", cluster.KubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig from path: %w", err)
		}
	} else {
		// Try in-cluster config (for on-prem clusters where manager is running)
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	}

	// Set timeout
	config.Timeout = 30 * time.Second

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return client, nil
}

// UpdateCluster updates cluster information
func (cm *ClusterManager) UpdateCluster(clusterID string, updates map[string]interface{}) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cluster, exists := cm.clusters[clusterID]
	if !exists {
		return fmt.Errorf("cluster not found: %s", clusterID)
	}

	// Update fields
	if name, ok := updates["name"].(string); ok && name != "" {
		cluster.Name = name
	}
	if provider, ok := updates["provider"].(string); ok {
		cluster.Provider = provider
	}
	if metadata, ok := updates["metadata"].(map[string]string); ok {
		if cluster.Metadata == nil {
			cluster.Metadata = make(map[string]string)
		}
		for k, v := range metadata {
			cluster.Metadata[k] = v
		}
	}

	// If kubeconfig changed, invalidate client cache
	if kubeconfig, ok := updates["kubeconfig"].(string); ok {
		cluster.Kubeconfig = kubeconfig
		delete(cm.clients, clusterID)
	}

	cluster.UpdatedAt = time.Now()

	return nil
}

// DeleteCluster removes a cluster
func (cm *ClusterManager) DeleteCluster(clusterID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.clusters[clusterID]; !exists {
		return fmt.Errorf("cluster not found: %s", clusterID)
	}

	delete(cm.clusters, clusterID)
	delete(cm.clients, clusterID)

	cm.logger.Info("deleted cluster", zap.String("cluster_id", clusterID))
	return nil
}

// TestConnection tests connection to a cluster
func (cm *ClusterManager) TestConnection(ctx context.Context, clusterID string) error {
	client, err := cm.GetClient(clusterID)
	if err != nil {
		return err
	}

	// Try to list namespaces as a connectivity test
	_, err = client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		cm.mu.Lock()
		if cluster, exists := cm.clusters[clusterID]; exists {
			cluster.Status = "error"
			cluster.UpdatedAt = time.Now()
		}
		cm.mu.Unlock()
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Update status
	cm.mu.Lock()
	if cluster, exists := cm.clusters[clusterID]; exists {
		cluster.Status = "connected"
		cluster.LastSeen = time.Now()
		cluster.UpdatedAt = time.Now()
	}
	cm.mu.Unlock()

	return nil
}

// RefreshAllConnections refreshes connections to all clusters
func (cm *ClusterManager) RefreshAllConnections(ctx context.Context) map[string]error {
	cm.mu.RLock()
	clusterIDs := make([]string, 0, len(cm.clusters))
	for id := range cm.clusters {
		clusterIDs = append(clusterIDs, id)
	}
	cm.mu.RUnlock()

	errors := make(map[string]error)
	for _, clusterID := range clusterIDs {
		if err := cm.TestConnection(ctx, clusterID); err != nil {
			errors[clusterID] = err
		}
	}

	return errors
}

