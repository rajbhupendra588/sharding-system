package operator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Operator manages automatic PostgreSQL shard provisioning
type Operator struct {
	client    kubernetes.Interface
	logger    *zap.Logger
	namespace string
	databases map[string]*ShardedDatabase
	mu        sync.RWMutex

	// Callbacks
	onShardReady func(dbName string, shard ShardInfo)
}

// NewOperator creates a new Kubernetes operator
func NewOperator(logger *zap.Logger, namespace string) (*Operator, error) {
	// Try in-cluster config first, then fall back to kubeconfig
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig for local development
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes config: %w", err)
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Operator{
		client:    client,
		logger:    logger,
		namespace: namespace,
		databases: make(map[string]*ShardedDatabase),
	}, nil
}

// NewOperatorWithClient creates an operator with a provided client (for testing)
func NewOperatorWithClient(client kubernetes.Interface, logger *zap.Logger, namespace string) *Operator {
	return &Operator{
		client:    client,
		logger:    logger,
		namespace: namespace,
		databases: make(map[string]*ShardedDatabase),
	}
}

// SetOnShardReady sets callback for when a shard becomes ready
func (o *Operator) SetOnShardReady(callback func(dbName string, shard ShardInfo)) {
	o.onShardReady = callback
}

// CreateShardedDatabase creates a new sharded database with automatic provisioning
func (o *Operator) CreateShardedDatabase(ctx context.Context, spec ShardedDatabaseSpec) (*ShardedDatabase, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Check if database already exists
	if _, exists := o.databases[spec.Name]; exists {
		return nil, fmt.Errorf("database %s already exists", spec.Name)
	}

	db := &ShardedDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: o.namespace,
		},
		Spec: spec,
		Status: ShardedDatabaseStatus{
			Phase:         "Creating",
			Shards:        make([]ShardInfo, 0, spec.ShardCount),
			CreatedAt:     time.Now(),
			SchemaVersion: 0,
		},
	}

	o.databases[spec.Name] = db

	// Create shards asynchronously
	go o.provisionShards(ctx, db)

	o.logger.Info("started creating sharded database",
		zap.String("name", spec.Name),
		zap.Int("shardCount", spec.ShardCount))

	return db, nil
}

// provisionShards creates all PostgreSQL shards for a database
func (o *Operator) provisionShards(ctx context.Context, db *ShardedDatabase) {
	var wg sync.WaitGroup
	errors := make(chan error, db.Spec.ShardCount)

	for i := 0; i < db.Spec.ShardCount; i++ {
		wg.Add(1)
		go func(shardIndex int) {
			defer wg.Done()
			if err := o.createShard(ctx, db, shardIndex); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if len(errs) > 0 {
		db.Status.Phase = "Failed"
		db.Status.Message = fmt.Sprintf("failed to create %d shards", len(errs))
		o.logger.Error("failed to create some shards",
			zap.String("database", db.Spec.Name),
			zap.Int("failedCount", len(errs)))
		return
	}

	// All shards created successfully
	now := time.Now()
	db.Status.Phase = "Ready"
	db.Status.ReadyAt = &now
	db.Status.ConnectionString = o.generateConnectionString(db)
	db.Status.ProxyEndpoint = fmt.Sprintf("sharding-proxy.%s.svc.cluster.local:6432", o.namespace)
	db.Status.Message = "All shards ready"

	o.logger.Info("sharded database ready",
		zap.String("name", db.Spec.Name),
		zap.Int("shardCount", len(db.Status.Shards)))
}

// createShard creates a single PostgreSQL shard
func (o *Operator) createShard(ctx context.Context, db *ShardedDatabase, index int) error {
	shardName := fmt.Sprintf("%s-shard-%d", db.Spec.Name, index)
	shardID := uuid.New().String()

	o.logger.Info("creating shard", zap.String("name", shardName), zap.Int("index", index))

	// Create PVC for persistent storage
	if err := o.createPVC(ctx, db, shardName); err != nil {
		return fmt.Errorf("failed to create PVC: %w", err)
	}

	// Create Secret for PostgreSQL credentials
	password := generatePassword()
	if err := o.createSecret(ctx, db, shardName, password); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	// Create StatefulSet for PostgreSQL
	if err := o.createStatefulSet(ctx, db, shardName, index); err != nil {
		return fmt.Errorf("failed to create StatefulSet: %w", err)
	}

	// Create Service for the shard
	if err := o.createService(ctx, db, shardName); err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Wait for pod to be ready
	if err := o.waitForPodReady(ctx, shardName); err != nil {
		return fmt.Errorf("pod failed to become ready: %w", err)
	}

	// Apply initial schema if provided
	if db.Spec.Schema != "" {
		if err := o.applySchema(ctx, db, shardName, db.Spec.Schema); err != nil {
			o.logger.Warn("failed to apply initial schema", zap.Error(err))
		}
	}

	// Record shard info
	shardInfo := ShardInfo{
		ID:        shardID,
		Name:      shardName,
		Host:      fmt.Sprintf("%s.%s.svc.cluster.local", shardName, o.namespace),
		Port:      5432,
		Database:  db.Spec.Name,
		Status:    "ready",
		PodName:   fmt.Sprintf("%s-0", shardName),
		PVCName:   fmt.Sprintf("data-%s-0", shardName),
		CreatedAt: time.Now(),
	}

	o.mu.Lock()
	db.Status.Shards = append(db.Status.Shards, shardInfo)
	o.mu.Unlock()

	// Notify callback
	if o.onShardReady != nil {
		o.onShardReady(db.Spec.Name, shardInfo)
	}

	o.logger.Info("shard created successfully", zap.String("name", shardName))
	return nil
}

// createPVC creates a PersistentVolumeClaim for the shard
func (o *Operator) createPVC(ctx context.Context, db *ShardedDatabase, shardName string) error {
	storageSize, err := resource.ParseQuantity(db.Spec.Storage.Size)
	if err != nil {
		return fmt.Errorf("invalid storage size: %w", err)
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("data-%s", shardName),
			Namespace: o.namespace,
			Labels: map[string]string{
				"app":      "sharding-system",
				"database": db.Spec.Name,
				"shard":    shardName,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageSize,
				},
			},
		},
	}

	if db.Spec.Storage.StorageClass != "" {
		pvc.Spec.StorageClassName = &db.Spec.Storage.StorageClass
	}

	_, err = o.client.CoreV1().PersistentVolumeClaims(o.namespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

// createSecret creates a Secret for PostgreSQL credentials
func (o *Operator) createSecret(ctx context.Context, db *ShardedDatabase, shardName, password string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-credentials", shardName),
			Namespace: o.namespace,
			Labels: map[string]string{
				"app":      "sharding-system",
				"database": db.Spec.Name,
				"shard":    shardName,
			},
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"POSTGRES_USER":     "sharding_admin",
			"POSTGRES_PASSWORD": password,
			"POSTGRES_DB":       db.Spec.Name,
		},
	}

	_, err := o.client.CoreV1().Secrets(o.namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

// createStatefulSet creates a StatefulSet for PostgreSQL
func (o *Operator) createStatefulSet(ctx context.Context, db *ShardedDatabase, shardName string, index int) error {
	replicas := int32(1)

	cpuLimit, _ := resource.ParseQuantity(db.Spec.Resources.CPU)
	memLimit, _ := resource.ParseQuantity(db.Spec.Resources.Memory)

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      shardName,
			Namespace: o.namespace,
			Labels: map[string]string{
				"app":         "sharding-system",
				"component":   "postgresql",
				"database":    db.Spec.Name,
				"shard":       shardName,
				"shard-index": fmt.Sprintf("%d", index),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: shardName,
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":   "sharding-system",
					"shard": shardName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":         "sharding-system",
						"component":   "postgresql",
						"database":    db.Spec.Name,
						"shard":       shardName,
						"shard-index": fmt.Sprintf("%d", index),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "postgresql",
							Image: "postgres:15-alpine",
							Ports: []corev1.ContainerPort{
								{
									Name:          "postgresql",
									ContainerPort: 5432,
								},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: fmt.Sprintf("%s-credentials", shardName),
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    cpuLimit,
									corev1.ResourceMemory: memLimit,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuLimit,
									corev1.ResourceMemory: memLimit,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/var/lib/postgresql/data",
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"pg_isready",
											"-U", "sharding_admin",
											"-d", db.Spec.Name,
										},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"pg_isready",
											"-U", "sharding_admin",
											"-d", db.Spec.Name,
										},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: fmt.Sprintf("data-%s", shardName),
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := o.client.AppsV1().StatefulSets(o.namespace).Create(ctx, sts, metav1.CreateOptions{})
	return err
}

// createService creates a headless Service for the shard
func (o *Operator) createService(ctx context.Context, db *ShardedDatabase, shardName string) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      shardName,
			Namespace: o.namespace,
			Labels: map[string]string{
				"app":      "sharding-system",
				"database": db.Spec.Name,
				"shard":    shardName,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":   "sharding-system",
				"shard": shardName,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "postgresql",
					Port:       5432,
					TargetPort: intstr.FromInt(5432),
				},
			},
			ClusterIP: corev1.ClusterIPNone, // Headless service
		},
	}

	_, err := o.client.CoreV1().Services(o.namespace).Create(ctx, svc, metav1.CreateOptions{})
	return err
}

// waitForPodReady waits for the PostgreSQL pod to be ready
func (o *Operator) waitForPodReady(ctx context.Context, shardName string) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	podName := fmt.Sprintf("%s-0", shardName)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for pod %s to be ready", podName)
		case <-ticker.C:
			pod, err := o.client.CoreV1().Pods(o.namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				continue // Pod might not exist yet
			}

			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
					return nil
				}
			}
		}
	}
}

// applySchema applies SQL schema to a shard
func (o *Operator) applySchema(ctx context.Context, db *ShardedDatabase, shardName, schema string) error {
	// Execute schema via kubectl exec or direct connection
	// For now, we'll use a Job to apply the schema
	o.logger.Info("applying schema to shard", zap.String("shard", shardName))
	// TODO: Implement schema application via Job or direct connection
	return nil
}

// generateConnectionString generates the proxy connection string
func (o *Operator) generateConnectionString(db *ShardedDatabase) string {
	return fmt.Sprintf("postgresql://sharding_admin@sharding-proxy.%s.svc.cluster.local:6432/%s?sslmode=disable",
		o.namespace, db.Spec.Name)
}

// GetDatabase retrieves a sharded database by name
func (o *Operator) GetDatabase(name string) (*ShardedDatabase, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	db, exists := o.databases[name]
	return db, exists
}

// ListDatabases returns all sharded databases
func (o *Operator) ListDatabases() []*ShardedDatabase {
	o.mu.RLock()
	defer o.mu.RUnlock()

	result := make([]*ShardedDatabase, 0, len(o.databases))
	for _, db := range o.databases {
		result = append(result, db)
	}
	return result
}

// DeleteDatabase deletes a sharded database and all its resources
func (o *Operator) DeleteDatabase(ctx context.Context, name string) error {
	o.mu.Lock()
	db, exists := o.databases[name]
	if !exists {
		o.mu.Unlock()
		return fmt.Errorf("database %s not found", name)
	}
	delete(o.databases, name)
	o.mu.Unlock()

	// Delete all shards
	for _, shard := range db.Status.Shards {
		if err := o.deleteShard(ctx, shard.Name); err != nil {
			o.logger.Warn("failed to delete shard", zap.String("shard", shard.Name), zap.Error(err))
		}
	}

	o.logger.Info("deleted sharded database", zap.String("name", name))
	return nil
}

// deleteShard deletes a single shard and its resources
func (o *Operator) deleteShard(ctx context.Context, shardName string) error {
	// Delete StatefulSet
	if err := o.client.AppsV1().StatefulSets(o.namespace).Delete(ctx, shardName, metav1.DeleteOptions{}); err != nil {
		o.logger.Warn("failed to delete StatefulSet", zap.String("name", shardName), zap.Error(err))
	}

	// Delete Service
	if err := o.client.CoreV1().Services(o.namespace).Delete(ctx, shardName, metav1.DeleteOptions{}); err != nil {
		o.logger.Warn("failed to delete Service", zap.String("name", shardName), zap.Error(err))
	}

	// Delete Secret
	secretName := fmt.Sprintf("%s-credentials", shardName)
	if err := o.client.CoreV1().Secrets(o.namespace).Delete(ctx, secretName, metav1.DeleteOptions{}); err != nil {
		o.logger.Warn("failed to delete Secret", zap.String("name", secretName), zap.Error(err))
	}

	// Delete PVC
	pvcName := fmt.Sprintf("data-%s", shardName)
	if err := o.client.CoreV1().PersistentVolumeClaims(o.namespace).Delete(ctx, pvcName, metav1.DeleteOptions{}); err != nil {
		o.logger.Warn("failed to delete PVC", zap.String("name", pvcName), zap.Error(err))
	}

	return nil
}

// ScaleShards adds or removes shards from a database
func (o *Operator) ScaleShards(ctx context.Context, name string, newCount int) error {
	o.mu.Lock()
	db, exists := o.databases[name]
	if !exists {
		o.mu.Unlock()
		return fmt.Errorf("database %s not found", name)
	}

	currentCount := len(db.Status.Shards)
	o.mu.Unlock()

	if newCount == currentCount {
		return nil
	}

	if newCount > currentCount {
		// Scale up - add new shards
		for i := currentCount; i < newCount; i++ {
			if err := o.createShard(ctx, db, i); err != nil {
				return fmt.Errorf("failed to create shard %d: %w", i, err)
			}
		}
	} else {
		// Scale down - remove shards (with data migration warning)
		o.logger.Warn("scaling down shards - data migration required",
			zap.String("database", name),
			zap.Int("from", currentCount),
			zap.Int("to", newCount))
		// TODO: Implement data migration before removing shards
	}

	o.mu.Lock()
	db.Spec.ShardCount = newCount
	o.mu.Unlock()

	return nil
}

// generatePassword generates a secure random password
func generatePassword() string {
	return uuid.New().String()[:16]
}

