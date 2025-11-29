package scanner

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sharding-system/pkg/models"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterManager manages multiple Kubernetes cluster connections
type ClusterManager struct {
	clusters map[string]*ClusterConnection
	mu       sync.RWMutex
	logger   *zap.Logger
}

// ClusterConnection represents a connection to a Kubernetes cluster
type ClusterConnection struct {
	Cluster   *models.Cluster
	Client    *kubernetes.Clientset
	Config    *rest.Config
	LastCheck time.Time
	Status    string
	Error     error
}

// NewClusterManager creates a new cluster manager
func NewClusterManager(logger *zap.Logger) *ClusterManager {
	return &ClusterManager{
		clusters: make(map[string]*ClusterConnection),
		logger:   logger,
	}
}

// RegisterCluster registers a new cluster for scanning
func (cm *ClusterManager) RegisterCluster(ctx context.Context, cluster *models.Cluster) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if cluster with same name already exists
	for _, conn := range cm.clusters {
		if conn.Cluster.Name == cluster.Name {
			return fmt.Errorf("cluster with name '%s' is already registered", cluster.Name)
		}
	}

	// Generate ID if not provided
	if cluster.ID == "" {
		cluster.ID = generateClusterID()
	}

	// Create K8s client
	config, clientset, err := cm.createK8sClient(cluster)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Test connection
	if err := cm.testConnection(ctx, clientset); err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	conn := &ClusterConnection{
		Cluster:   cluster,
		Client:    clientset,
		Config:    config,
		LastCheck: time.Now(),
		Status:    "active",
	}

	cm.clusters[cluster.ID] = conn
	cm.logger.Info("registered cluster", zap.String("cluster_id", cluster.ID), zap.String("name", cluster.Name))

	return nil
}

// UnregisterCluster removes a cluster
func (cm *ClusterManager) UnregisterCluster(clusterID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.clusters, clusterID)
	cm.logger.Info("unregistered cluster", zap.String("cluster_id", clusterID))
}

// GetCluster gets a cluster connection by ID
func (cm *ClusterManager) GetCluster(clusterID string) (*ClusterConnection, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, ok := cm.clusters[clusterID]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", clusterID)
	}

	return conn, nil
}

// ListClusters returns all registered clusters
func (cm *ClusterManager) ListClusters() []*models.Cluster {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clusters := make([]*models.Cluster, 0, len(cm.clusters))
	for _, conn := range cm.clusters {
		clusters = append(clusters, conn.Cluster)
	}
	return clusters
}

// createK8sClient creates a Kubernetes client from cluster config
func (cm *ClusterManager) createK8sClient(cluster *models.Cluster) (*rest.Config, *kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Try different methods to get kubeconfig
	if cluster.Kubeconfig != "" {
		// Check if it's a file path or base64 encoded
		if _, err := os.Stat(cluster.Kubeconfig); err == nil {
			// It's a file path
			config, err = clientcmd.BuildConfigFromFlags("", cluster.Kubeconfig)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load kubeconfig from file: %w", err)
			}
		} else {
			// Try as base64 encoded kubeconfig
			kubeconfigData, err := base64.StdEncoding.DecodeString(cluster.Kubeconfig)
			if err == nil {
				// Parse the kubeconfig
				clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfigData)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
				}
				config, err = clientConfig.ClientConfig()
				if err != nil {
					return nil, nil, fmt.Errorf("failed to get client config: %w", err)
				}
			} else {
				// Try as direct kubeconfig content
				clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(cluster.Kubeconfig))
				if err != nil {
					return nil, nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
				}
				config, err = clientConfig.ClientConfig()
				if err != nil {
					return nil, nil, fmt.Errorf("failed to get client config: %w", err)
				}
			}
		}

		// Use specific context if provided
		if cluster.Context != "" {
			// Load kubeconfig and switch context
			var kubeconfigPath string
			if _, err := os.Stat(cluster.Kubeconfig); err == nil {
				kubeconfigPath = cluster.Kubeconfig
			} else {
				// Use default kubeconfig location
				kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
			}

			configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
				&clientcmd.ConfigOverrides{CurrentContext: cluster.Context},
			)
			config, err = configLoader.ClientConfig()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load context: %w", err)
			}
		}
	} else if cluster.Endpoint != "" {
		// Direct endpoint connection (for cloud providers)
		config = &rest.Config{
			Host: cluster.Endpoint,
		}

		// Add authentication if credentials provided
		if token, ok := cluster.Credentials["token"]; ok {
			config.BearerToken = token
		}
		if cert, ok := cluster.Credentials["cert"]; ok {
			config.CertData = []byte(cert)
		}
		if key, ok := cluster.Credentials["key"]; ok {
			config.KeyData = []byte(key)
		}
		if ca, ok := cluster.Credentials["ca"]; ok {
			config.CAData = []byte(ca)
		}
	} else {
		// Try in-cluster config first
		config, err = rest.InClusterConfig()
		if err != nil {
			// Fall back to default kubeconfig
			config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
			}
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return config, clientset, nil
}

// testConnection tests the connection to a cluster
func (cm *ClusterManager) testConnection(ctx context.Context, clientset *kubernetes.Clientset) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	return err
}

// generateClusterID generates a unique cluster ID
func generateClusterID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

