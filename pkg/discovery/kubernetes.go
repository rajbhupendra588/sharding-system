package discovery

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// DiscoveredApp represents an application discovered in Kubernetes
type DiscoveredApp struct {
	Namespace    string            `json:"namespace"`
	Name         string            `json:"name"`
	Type         string            `json:"type"` // "deployment", "statefulset", "pod"
	DatabaseName string            `json:"database_name"`
	DatabaseURL  string            `json:"database_url,omitempty"`
	DatabaseHost string            `json:"database_host,omitempty"`
	DatabasePort string            `json:"database_port,omitempty"`
	DatabaseUser string            `json:"database_user,omitempty"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	IsRegistered bool              `json:"is_registered"` // Whether already registered as client app
}

// KubernetesDiscovery discovers applications and databases in Kubernetes clusters
type KubernetesDiscovery struct {
	client         *kubernetes.Clientset
	logger         *zap.Logger
	registeredApps map[string]bool // Track registered app names
}

// NewKubernetesDiscovery creates a new Kubernetes discovery service
func NewKubernetesDiscovery(logger *zap.Logger, registeredAppNames []string) (*KubernetesDiscovery, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first (when running in K8s)
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file (for local development)
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	registeredMap := make(map[string]bool)
	for _, name := range registeredAppNames {
		registeredMap[name] = true
	}

	return &KubernetesDiscovery{
		client:         clientset,
		logger:         logger,
		registeredApps: registeredMap,
	}, nil
}

// NewKubernetesDiscoveryFromClient creates a new Kubernetes discovery service from an existing client
func NewKubernetesDiscoveryFromClient(client *kubernetes.Clientset, logger *zap.Logger, registeredAppNames []string) (*KubernetesDiscovery, error) {
	registeredMap := make(map[string]bool)
	for _, name := range registeredAppNames {
		registeredMap[name] = true
	}

	return &KubernetesDiscovery{
		client:         client,
		logger:         logger,
		registeredApps: registeredMap,
	}, nil
}

// DiscoverApplications discovers all applications in Kubernetes namespaces
func (k *KubernetesDiscovery) DiscoverApplications(ctx context.Context) ([]DiscoveredApp, error) {
	// Initialize as empty slice (not nil) to ensure JSON encoding produces [] not null
	discoveredApps := make([]DiscoveredApp, 0)

	// Get all namespaces
	namespaces, err := k.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	for _, ns := range namespaces.Items {
		// Skip system namespaces
		if k.isSystemNamespace(ns.Name) {
			continue
		}

		// Discover deployments
		deployments, err := k.client.AppsV1().Deployments(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			k.logger.Warn("failed to list deployments", zap.String("namespace", ns.Name), zap.Error(err))
			continue
		}

		for _, deployment := range deployments.Items {
			app := k.discoverFromDeployment(ctx, &deployment)
			if app != nil {
				discoveredApps = append(discoveredApps, *app)
			}
		}

		// Discover StatefulSets
		statefulSets, err := k.client.AppsV1().StatefulSets(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			k.logger.Warn("failed to list statefulsets", zap.String("namespace", ns.Name), zap.Error(err))
			continue
		}

		for _, sts := range statefulSets.Items {
			app := k.discoverFromStatefulSet(ctx, &sts)
			if app != nil {
				discoveredApps = append(discoveredApps, *app)
			}
		}
	}

	// Ensure we always return a non-nil slice
	if discoveredApps == nil {
		discoveredApps = make([]DiscoveredApp, 0)
	}

	return discoveredApps, nil
}

// discoverFromDeployment extracts application and database info from a deployment
func (k *KubernetesDiscovery) discoverFromDeployment(ctx context.Context, deployment *appsv1.Deployment) *DiscoveredApp {
	app := &DiscoveredApp{
		Namespace:    deployment.Namespace,
		Name:         deployment.Name,
		Type:         "deployment",
		Labels:       deployment.Labels,
		Annotations:  deployment.Annotations,
		IsRegistered: k.registeredApps[deployment.Name],
	}

	// Extract database info from pod template
	k.extractDatabaseInfo(&deployment.Spec.Template.Spec, app)

	// If no database found, try to infer from name
	if app.DatabaseName == "" {
		app.DatabaseName = k.inferDatabaseName(deployment.Name, deployment.Namespace)
	}

	return app
}

// discoverFromStatefulSet extracts application and database info from a statefulset
func (k *KubernetesDiscovery) discoverFromStatefulSet(ctx context.Context, sts *appsv1.StatefulSet) *DiscoveredApp {
	app := &DiscoveredApp{
		Namespace:    sts.Namespace,
		Name:         sts.Name,
		Type:         "statefulset",
		Labels:       sts.Labels,
		Annotations:  sts.Annotations,
		IsRegistered: k.registeredApps[sts.Name],
	}

	// Extract database info from pod template
	k.extractDatabaseInfo(&sts.Spec.Template.Spec, app)

	// If no database found, try to infer from name
	if app.DatabaseName == "" {
		app.DatabaseName = k.inferDatabaseName(sts.Name, sts.Namespace)
	}

	return app
}

// extractDatabaseInfo extracts database connection info from pod spec
func (k *KubernetesDiscovery) extractDatabaseInfo(podSpec *corev1.PodSpec, app *DiscoveredApp) {
	// Check environment variables in all containers
	for _, container := range podSpec.Containers {
		for _, env := range container.Env {
			// Check for database URL patterns
			if strings.HasPrefix(env.Name, "DATABASE_URL") || strings.HasPrefix(env.Name, "DB_URL") {
				if env.Value != "" {
					app.DatabaseURL = env.Value
					k.parseDatabaseURL(env.Value, app)
				} else if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
					// Database URL is in a secret - mark that we need to fetch it
					app.DatabaseURL = fmt.Sprintf("secret:%s/%s", env.ValueFrom.SecretKeyRef.Name, env.ValueFrom.SecretKeyRef.Key)
				}
			}

			// Check for individual database connection components
			if env.Name == "DB_HOST" || env.Name == "DATABASE_HOST" {
				if env.Value != "" {
					app.DatabaseHost = env.Value
				}
			}
			if env.Name == "DB_PORT" || env.Name == "DATABASE_PORT" {
				if env.Value != "" {
					app.DatabasePort = env.Value
				}
			}
			if env.Name == "DB_NAME" || env.Name == "DATABASE_NAME" {
				if env.Value != "" {
					app.DatabaseName = env.Value
				}
			}
			if env.Name == "DB_USER" || env.Name == "DATABASE_USER" {
				if env.Value != "" {
					app.DatabaseUser = env.Value
				}
			}
		}
	}

	// Check ConfigMaps referenced in volumes
	for _, volume := range podSpec.Volumes {
		if volume.ConfigMap != nil {
			// Try to get configmap and check for database config
			// This would require additional API call, so we'll skip for now
			// but mark that configmap might contain DB info
		}
	}
}

// parseDatabaseURL parses a database URL and extracts components
func (k *KubernetesDiscovery) parseDatabaseURL(url string, app *DiscoveredApp) {
	// PostgreSQL URL pattern: postgres://user:pass@host:port/dbname
	postgresRegex := regexp.MustCompile(`postgres(ql)?://([^:]+):([^@]+)@([^:]+):(\d+)/(.+)`)
	matches := postgresRegex.FindStringSubmatch(url)
	if len(matches) == 7 {
		app.DatabaseUser = matches[2]
		app.DatabaseHost = matches[4]
		app.DatabasePort = matches[5]
		app.DatabaseName = matches[6]
		return
	}

	// MySQL URL pattern: mysql://user:pass@host:port/dbname
	mysqlRegex := regexp.MustCompile(`mysql://([^:]+):([^@]+)@([^:]+):(\d+)/(.+)`)
	matches = mysqlRegex.FindStringSubmatch(url)
	if len(matches) == 6 {
		app.DatabaseUser = matches[1]
		app.DatabaseHost = matches[3]
		app.DatabasePort = matches[4]
		app.DatabaseName = matches[5]
		return
	}
}

// inferDatabaseName tries to infer database name from application name
func (k *KubernetesDiscovery) inferDatabaseName(appName, namespace string) string {
	// Remove common suffixes
	name := strings.ToLower(appName)
	name = strings.TrimSuffix(name, "-deployment")
	name = strings.TrimSuffix(name, "-app")
	name = strings.TrimSuffix(name, "-service")
	name = strings.TrimSuffix(name, "-api")

	// Add _db suffix if not present
	if !strings.HasSuffix(name, "_db") && !strings.HasSuffix(name, "-db") {
		name = name + "_db"
	}

	return name
}

// isSystemNamespace checks if a namespace is a system namespace
func (k *KubernetesDiscovery) isSystemNamespace(name string) bool {
	systemNamespaces := []string{
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"default", // Optionally skip default
	}
	for _, sysNs := range systemNamespaces {
		if name == sysNs {
			return true
		}
	}
	return false
}

// UpdateRegisteredApps updates the list of registered application names
func (k *KubernetesDiscovery) UpdateRegisteredApps(names []string) {
	k.registeredApps = make(map[string]bool)
	for _, name := range names {
		k.registeredApps[name] = true
	}
}
