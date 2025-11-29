package discovery

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// WatchCallback is called when resources are discovered or updated
type WatchCallback func(apps []DiscoveredApp)

// ConfigWatchCallback is called when configuration changes
type ConfigWatchCallback func(namespace, name string, data map[string]string)

// KubernetesWatcher watches Kubernetes resources for changes in real-time
type KubernetesWatcher struct {
	client           *kubernetes.Clientset
	logger           *zap.Logger
	registeredApps   map[string]bool
	mu               sync.RWMutex
	stopCh           chan struct{}
	callbacks        []WatchCallback
	configCallbacks  []ConfigWatchCallback
	discoveredApps   map[string]*DiscoveredApp // keyed by namespace/name
	watchNamespaces  []string                  // namespaces to watch, empty for all
	labelSelector    string                    // label selector for filtering
	resyncInterval   time.Duration
}

// WatcherConfig holds configuration for the watcher
type WatcherConfig struct {
	Namespaces     []string
	LabelSelector  string
	ResyncInterval time.Duration
}

// NewKubernetesWatcher creates a new Kubernetes watcher
func NewKubernetesWatcher(logger *zap.Logger, cfg WatcherConfig) (*KubernetesWatcher, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first (when running in K8s)
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file (for local development)
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	resync := cfg.ResyncInterval
	if resync == 0 {
		resync = 30 * time.Second
	}

	return &KubernetesWatcher{
		client:          clientset,
		logger:          logger,
		registeredApps:  make(map[string]bool),
		stopCh:          make(chan struct{}),
		callbacks:       make([]WatchCallback, 0),
		configCallbacks: make([]ConfigWatchCallback, 0),
		discoveredApps:  make(map[string]*DiscoveredApp),
		watchNamespaces: cfg.Namespaces,
		labelSelector:   cfg.LabelSelector,
		resyncInterval:  resync,
	}, nil
}

// OnDiscovery registers a callback for discovery events
func (w *KubernetesWatcher) OnDiscovery(cb WatchCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, cb)
}

// OnConfigChange registers a callback for config changes
func (w *KubernetesWatcher) OnConfigChange(cb ConfigWatchCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.configCallbacks = append(w.configCallbacks, cb)
}

// Start starts watching Kubernetes resources
func (w *KubernetesWatcher) Start(ctx context.Context) {
	w.logger.Info("starting Kubernetes watcher")

	// Watch deployments
	go w.watchDeployments(ctx)

	// Watch StatefulSets
	go w.watchStatefulSets(ctx)

	// Watch ConfigMaps for config hot-reload
	go w.watchConfigMaps(ctx)

	// Periodic resync to catch any missed events
	go w.periodicResync(ctx)
}

// Stop stops the watcher
func (w *KubernetesWatcher) Stop() {
	close(w.stopCh)
	w.logger.Info("Kubernetes watcher stopped")
}

// watchDeployments watches deployment changes
func (w *KubernetesWatcher) watchDeployments(ctx context.Context) {
	for {
		namespaces := w.getWatchNamespaces(ctx)
		for _, ns := range namespaces {
			go w.watchDeploymentsInNamespace(ctx, ns)
		}

		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-time.After(5 * time.Minute):
			// Refresh namespace list periodically
		}
	}
}

// watchDeploymentsInNamespace watches deployments in a specific namespace
func (w *KubernetesWatcher) watchDeploymentsInNamespace(ctx context.Context, namespace string) {
	opts := metav1.ListOptions{}
	if w.labelSelector != "" {
		opts.LabelSelector = w.labelSelector
	}

	watcher, err := w.client.AppsV1().Deployments(namespace).Watch(ctx, opts)
	if err != nil {
		w.logger.Error("failed to watch deployments", zap.String("namespace", namespace), zap.Error(err))
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				w.logger.Warn("deployment watcher closed, restarting", zap.String("namespace", namespace))
				return
			}
			w.handleDeploymentEvent(ctx, event)
		}
	}
}

// handleDeploymentEvent handles a deployment watch event
func (w *KubernetesWatcher) handleDeploymentEvent(ctx context.Context, event watch.Event) {
	deployment, ok := event.Object.(*appsv1.Deployment)
	if !ok {
		return
	}

	key := deployment.Namespace + "/" + deployment.Name

	switch event.Type {
	case watch.Added, watch.Modified:
		app := w.discoverFromDeployment(deployment)
		if app != nil {
			w.mu.Lock()
			w.discoveredApps[key] = app
			w.mu.Unlock()
			w.notifyCallbacks()
		}
	case watch.Deleted:
		w.mu.Lock()
		delete(w.discoveredApps, key)
		w.mu.Unlock()
		w.notifyCallbacks()
	}
}

// watchStatefulSets watches StatefulSet changes
func (w *KubernetesWatcher) watchStatefulSets(ctx context.Context) {
	for {
		namespaces := w.getWatchNamespaces(ctx)
		for _, ns := range namespaces {
			go w.watchStatefulSetsInNamespace(ctx, ns)
		}

		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-time.After(5 * time.Minute):
			// Refresh namespace list periodically
		}
	}
}

// watchStatefulSetsInNamespace watches StatefulSets in a specific namespace
func (w *KubernetesWatcher) watchStatefulSetsInNamespace(ctx context.Context, namespace string) {
	opts := metav1.ListOptions{}
	if w.labelSelector != "" {
		opts.LabelSelector = w.labelSelector
	}

	watcher, err := w.client.AppsV1().StatefulSets(namespace).Watch(ctx, opts)
	if err != nil {
		w.logger.Error("failed to watch statefulsets", zap.String("namespace", namespace), zap.Error(err))
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				w.logger.Warn("statefulset watcher closed, restarting", zap.String("namespace", namespace))
				return
			}
			w.handleStatefulSetEvent(ctx, event)
		}
	}
}

// handleStatefulSetEvent handles a StatefulSet watch event
func (w *KubernetesWatcher) handleStatefulSetEvent(ctx context.Context, event watch.Event) {
	sts, ok := event.Object.(*appsv1.StatefulSet)
	if !ok {
		return
	}

	key := sts.Namespace + "/" + sts.Name

	switch event.Type {
	case watch.Added, watch.Modified:
		app := w.discoverFromStatefulSet(sts)
		if app != nil {
			w.mu.Lock()
			w.discoveredApps[key] = app
			w.mu.Unlock()
			w.notifyCallbacks()
		}
	case watch.Deleted:
		w.mu.Lock()
		delete(w.discoveredApps, key)
		w.mu.Unlock()
		w.notifyCallbacks()
	}
}

// watchConfigMaps watches ConfigMap changes for hot-reload
func (w *KubernetesWatcher) watchConfigMaps(ctx context.Context) {
	namespace := "sharding-system"
	if len(w.watchNamespaces) > 0 {
		namespace = w.watchNamespaces[0]
	}

	opts := metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/part-of=sharding-system",
	}

	watcher, err := w.client.CoreV1().ConfigMaps(namespace).Watch(ctx, opts)
	if err != nil {
		w.logger.Error("failed to watch configmaps", zap.Error(err))
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				w.logger.Warn("configmap watcher closed")
				// Restart watcher
				go w.watchConfigMaps(ctx)
				return
			}
			w.handleConfigMapEvent(event)
		}
	}
}

// handleConfigMapEvent handles ConfigMap changes
func (w *KubernetesWatcher) handleConfigMapEvent(event watch.Event) {
	cm, ok := event.Object.(*corev1.ConfigMap)
	if !ok {
		return
	}

	switch event.Type {
	case watch.Modified:
		w.logger.Info("configmap changed", zap.String("namespace", cm.Namespace), zap.String("name", cm.Name))
		w.mu.RLock()
		callbacks := w.configCallbacks
		w.mu.RUnlock()

		for _, cb := range callbacks {
			cb(cm.Namespace, cm.Name, cm.Data)
		}
	}
}

// periodicResync periodically resyncs discovered apps
func (w *KubernetesWatcher) periodicResync(ctx context.Context) {
	ticker := time.NewTicker(w.resyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.resync(ctx)
		}
	}
}

// resync performs a full resync of discovered apps
func (w *KubernetesWatcher) resync(ctx context.Context) {
	w.logger.Debug("resyncing discovered applications")

	namespaces := w.getWatchNamespaces(ctx)
	newApps := make(map[string]*DiscoveredApp)

	for _, ns := range namespaces {
		// List deployments
		deployments, err := w.client.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{
			LabelSelector: w.labelSelector,
		})
		if err != nil {
			w.logger.Warn("failed to list deployments during resync", zap.String("namespace", ns), zap.Error(err))
			continue
		}

		for _, deployment := range deployments.Items {
			app := w.discoverFromDeployment(&deployment)
			if app != nil {
				key := deployment.Namespace + "/" + deployment.Name
				newApps[key] = app
			}
		}

		// List StatefulSets
		statefulSets, err := w.client.AppsV1().StatefulSets(ns).List(ctx, metav1.ListOptions{
			LabelSelector: w.labelSelector,
		})
		if err != nil {
			w.logger.Warn("failed to list statefulsets during resync", zap.String("namespace", ns), zap.Error(err))
			continue
		}

		for _, sts := range statefulSets.Items {
			app := w.discoverFromStatefulSet(&sts)
			if app != nil {
				key := sts.Namespace + "/" + sts.Name
				newApps[key] = app
			}
		}
	}

	w.mu.Lock()
	w.discoveredApps = newApps
	w.mu.Unlock()

	w.notifyCallbacks()
}

// getWatchNamespaces returns namespaces to watch
func (w *KubernetesWatcher) getWatchNamespaces(ctx context.Context) []string {
	if len(w.watchNamespaces) > 0 {
		return w.watchNamespaces
	}

	// Get all non-system namespaces
	namespaces, err := w.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		w.logger.Error("failed to list namespaces", zap.Error(err))
		return []string{"default"}
	}

	result := make([]string, 0)
	systemNamespaces := map[string]bool{
		"kube-system":     true,
		"kube-public":     true,
		"kube-node-lease": true,
	}

	for _, ns := range namespaces.Items {
		if !systemNamespaces[ns.Name] {
			result = append(result, ns.Name)
		}
	}

	return result
}

// notifyCallbacks notifies all registered callbacks
func (w *KubernetesWatcher) notifyCallbacks() {
	w.mu.RLock()
	apps := make([]DiscoveredApp, 0, len(w.discoveredApps))
	for _, app := range w.discoveredApps {
		apps = append(apps, *app)
	}
	callbacks := w.callbacks
	w.mu.RUnlock()

	for _, cb := range callbacks {
		cb(apps)
	}
}

// discoverFromDeployment extracts app info from deployment
func (w *KubernetesWatcher) discoverFromDeployment(deployment *appsv1.Deployment) *DiscoveredApp {
	w.mu.RLock()
	isRegistered := w.registeredApps[deployment.Name]
	w.mu.RUnlock()

	app := &DiscoveredApp{
		Namespace:    deployment.Namespace,
		Name:         deployment.Name,
		Type:         "deployment",
		Labels:       deployment.Labels,
		Annotations:  deployment.Annotations,
		IsRegistered: isRegistered,
	}

	// Extract database info from pod template
	extractDatabaseInfoFromPodSpec(&deployment.Spec.Template.Spec, app)

	// Infer database name if not found
	if app.DatabaseName == "" {
		app.DatabaseName = inferDatabaseName(deployment.Name)
	}

	return app
}

// discoverFromStatefulSet extracts app info from StatefulSet
func (w *KubernetesWatcher) discoverFromStatefulSet(sts *appsv1.StatefulSet) *DiscoveredApp {
	w.mu.RLock()
	isRegistered := w.registeredApps[sts.Name]
	w.mu.RUnlock()

	app := &DiscoveredApp{
		Namespace:    sts.Namespace,
		Name:         sts.Name,
		Type:         "statefulset",
		Labels:       sts.Labels,
		Annotations:  sts.Annotations,
		IsRegistered: isRegistered,
	}

	// Extract database info from pod template
	extractDatabaseInfoFromPodSpec(&sts.Spec.Template.Spec, app)

	// Infer database name if not found
	if app.DatabaseName == "" {
		app.DatabaseName = inferDatabaseName(sts.Name)
	}

	return app
}

// UpdateRegisteredApps updates the list of registered apps
func (w *KubernetesWatcher) UpdateRegisteredApps(names []string) {
	w.mu.Lock()
	w.registeredApps = make(map[string]bool)
	for _, name := range names {
		w.registeredApps[name] = true
	}
	w.mu.Unlock()
}

// GetDiscoveredApps returns all discovered apps
func (w *KubernetesWatcher) GetDiscoveredApps() []DiscoveredApp {
	w.mu.RLock()
	defer w.mu.RUnlock()

	apps := make([]DiscoveredApp, 0, len(w.discoveredApps))
	for _, app := range w.discoveredApps {
		apps = append(apps, *app)
	}
	return apps
}

// extractDatabaseInfoFromPodSpec extracts database info from pod spec
func extractDatabaseInfoFromPodSpec(podSpec *corev1.PodSpec, app *DiscoveredApp) {
	for _, container := range podSpec.Containers {
		for _, env := range container.Env {
			switch env.Name {
			case "DATABASE_URL", "DB_URL":
				if env.Value != "" {
					app.DatabaseURL = env.Value
				}
			case "DATABASE_HOST", "DB_HOST":
				if env.Value != "" {
					app.DatabaseHost = env.Value
				}
			case "DATABASE_PORT", "DB_PORT":
				if env.Value != "" {
					app.DatabasePort = env.Value
				}
			case "DATABASE_NAME", "DB_NAME":
				if env.Value != "" {
					app.DatabaseName = env.Value
				}
			case "DATABASE_USER", "DB_USER":
				if env.Value != "" {
					app.DatabaseUser = env.Value
				}
			}
		}
	}
}

// inferDatabaseName infers database name from app name
func inferDatabaseName(appName string) string {
	return appName + "_db"
}

