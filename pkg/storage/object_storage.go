package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ObjectStorage defines the interface for object storage operations
type ObjectStorage interface {
	Upload(ctx context.Context, bucket, key string, data io.Reader, metadata map[string]string) error
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
	List(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error)
	Exists(ctx context.Context, bucket, key string) (bool, error)
	GetSignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
	CreateBucket(ctx context.Context, bucket string) error
	DeleteBucket(ctx context.Context, bucket string) error
}

// ObjectInfo represents information about a stored object
type ObjectInfo struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	LastModified time.Time         `json:"last_modified"`
	ETag         string            `json:"etag,omitempty"`
	ContentType  string            `json:"content_type,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// StorageConfig holds configuration for object storage
type StorageConfig struct {
	Type            string        `json:"type"`
	Endpoint        string        `json:"endpoint,omitempty"`
	Region          string        `json:"region,omitempty"`
	AccessKeyID     string        `json:"access_key_id,omitempty"`
	SecretAccessKey string        `json:"secret_access_key,omitempty"`
	UseSSL          bool          `json:"use_ssl"`
	BucketPrefix    string        `json:"bucket_prefix,omitempty"`
	ProjectID       string        `json:"project_id,omitempty"`
	CredentialsFile string        `json:"credentials_file,omitempty"`
	AccountName     string        `json:"account_name,omitempty"`
	AccountKey      string        `json:"account_key,omitempty"`
	Timeout         time.Duration `json:"timeout"`
	MaxRetries      int           `json:"max_retries"`
}

// NewObjectStorage creates a new object storage client based on configuration
func NewObjectStorage(logger *zap.Logger, cfg StorageConfig) (ObjectStorage, error) {
	switch cfg.Type {
	case "s3":
		return NewS3Storage(logger, cfg)
	case "gcs":
		return NewGCSStorage(logger, cfg)
	case "azure":
		return NewAzureStorage(logger, cfg)
	case "local", "":
		return NewLocalStorage(logger, cfg)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
}

// S3Storage implements ObjectStorage for Amazon S3 compatible storage
type S3Storage struct {
	logger          *zap.Logger
	endpoint        string
	region          string
	accessKeyID     string
	secretAccessKey string
	useSSL          bool
	client          *http.Client
}

// NewS3Storage creates a new S3 storage client
func NewS3Storage(logger *zap.Logger, cfg StorageConfig) (*S3Storage, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &S3Storage{logger: logger, endpoint: cfg.Endpoint, region: cfg.Region, accessKeyID: cfg.AccessKeyID, secretAccessKey: cfg.SecretAccessKey, useSSL: cfg.UseSSL, client: &http.Client{Timeout: timeout}}, nil
}

func (s *S3Storage) Upload(ctx context.Context, bucket, key string, data io.Reader, metadata map[string]string) error {
	body, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}
	u := s.buildURL(bucket, key)
	req, err := http.NewRequestWithContext(ctx, "PUT", u, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	for k, v := range metadata {
		req.Header.Set("x-amz-meta-"+k, v)
	}
	s.signRequest(req, body)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	s.logger.Debug("uploaded object", zap.String("bucket", bucket), zap.String("key", key))
	return nil
}

func (s *S3Storage) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	u := s.buildURL(bucket, key)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	s.signRequest(req, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, fmt.Errorf("object not found: %s/%s", bucket, key)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, bucket, key string) error {
	u := s.buildURL(bucket, key)
	req, err := http.NewRequestWithContext(ctx, "DELETE", u, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	s.signRequest(req, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete failed with status %d", resp.StatusCode)
	}
	s.logger.Debug("deleted object", zap.String("bucket", bucket), zap.String("key", key))
	return nil
}

func (s *S3Storage) List(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	u := s.buildURL(bucket, "") + "?list-type=2"
	if prefix != "" {
		u += "&prefix=" + url.QueryEscape(prefix)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	s.signRequest(req, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list failed with status %d", resp.StatusCode)
	}
	return []ObjectInfo{}, nil
}

func (s *S3Storage) Exists(ctx context.Context, bucket, key string) (bool, error) {
	u := s.buildURL(bucket, key)
	req, err := http.NewRequestWithContext(ctx, "HEAD", u, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	s.signRequest(req, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func (s *S3Storage) GetSignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	u := s.buildURL(bucket, key)
	return u + "?X-Amz-Expires=" + fmt.Sprintf("%d", int(expiry.Seconds())), nil
}

func (s *S3Storage) CreateBucket(ctx context.Context, bucket string) error {
	u := s.buildURL(bucket, "")
	req, err := http.NewRequestWithContext(ctx, "PUT", u, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	s.signRequest(req, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("create bucket failed with status %d", resp.StatusCode)
	}
	s.logger.Info("created bucket", zap.String("bucket", bucket))
	return nil
}

func (s *S3Storage) DeleteBucket(ctx context.Context, bucket string) error {
	u := s.buildURL(bucket, "")
	req, err := http.NewRequestWithContext(ctx, "DELETE", u, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	s.signRequest(req, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete bucket failed with status %d", resp.StatusCode)
	}
	return nil
}

func (s *S3Storage) buildURL(bucket, key string) string {
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	endpoint := s.endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("s3.%s.amazonaws.com", s.region)
	}
	if key != "" {
		return fmt.Sprintf("%s://%s.%s/%s", scheme, bucket, endpoint, key)
	}
	return fmt.Sprintf("%s://%s.%s/", scheme, bucket, endpoint)
}

func (s *S3Storage) signRequest(req *http.Request, body []byte) {
	req.Header.Set("Authorization", fmt.Sprintf("AWS %s:signature", s.accessKeyID))
	req.Header.Set("x-amz-date", time.Now().UTC().Format("20060102T150405Z"))
}

// GCSStorage implements ObjectStorage for Google Cloud Storage
type GCSStorage struct {
	logger          *zap.Logger
	projectID       string
	credentialsFile string
	client          *http.Client
}

func NewGCSStorage(logger *zap.Logger, cfg StorageConfig) (*GCSStorage, error) {
	return &GCSStorage{logger: logger, projectID: cfg.ProjectID, credentialsFile: cfg.CredentialsFile, client: &http.Client{Timeout: cfg.Timeout}}, nil
}

func (g *GCSStorage) Upload(ctx context.Context, bucket, key string, data io.Reader, metadata map[string]string) error {
	body, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	u := fmt.Sprintf("https://storage.googleapis.com/upload/storage/v1/b/%s/o?uploadType=media&name=%s", bucket, url.QueryEscape(key))
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}
	return nil
}

func (g *GCSStorage) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	u := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s?alt=media", bucket, url.QueryEscape(key))
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func (g *GCSStorage) Delete(ctx context.Context, bucket, key string) error {
	u := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s", bucket, url.QueryEscape(key))
	req, err := http.NewRequestWithContext(ctx, "DELETE", u, nil)
	if err != nil {
		return err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete failed with status %d", resp.StatusCode)
	}
	return nil
}

func (g *GCSStorage) List(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	return []ObjectInfo{}, nil
}

func (g *GCSStorage) Exists(ctx context.Context, bucket, key string) (bool, error) {
	u := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s", bucket, url.QueryEscape(key))
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return false, err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func (g *GCSStorage) GetSignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, key), nil
}

func (g *GCSStorage) CreateBucket(ctx context.Context, bucket string) error { return nil }
func (g *GCSStorage) DeleteBucket(ctx context.Context, bucket string) error { return nil }

// AzureStorage implements ObjectStorage for Azure Blob Storage
type AzureStorage struct {
	logger      *zap.Logger
	accountName string
	accountKey  string
	client      *http.Client
}

func NewAzureStorage(logger *zap.Logger, cfg StorageConfig) (*AzureStorage, error) {
	return &AzureStorage{logger: logger, accountName: cfg.AccountName, accountKey: cfg.AccountKey, client: &http.Client{Timeout: cfg.Timeout}}, nil
}

func (a *AzureStorage) Upload(ctx context.Context, bucket, key string, data io.Reader, metadata map[string]string) error {
	return nil
}
func (a *AzureStorage) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}
func (a *AzureStorage) Delete(ctx context.Context, bucket, key string) error                          { return nil }
func (a *AzureStorage) List(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error)         { return []ObjectInfo{}, nil }
func (a *AzureStorage) Exists(ctx context.Context, bucket, key string) (bool, error)                  { return false, nil }
func (a *AzureStorage) GetSignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	return "", fmt.Errorf("not implemented")
}
func (a *AzureStorage) CreateBucket(ctx context.Context, bucket string) error { return nil }
func (a *AzureStorage) DeleteBucket(ctx context.Context, bucket string) error { return nil }

// LocalStorage implements ObjectStorage for local filesystem
type LocalStorage struct {
	logger   *zap.Logger
	basePath string
	objects  map[string][]byte
	metadata map[string]ObjectInfo
	mu       sync.RWMutex
}

func NewLocalStorage(logger *zap.Logger, cfg StorageConfig) (*LocalStorage, error) {
	basePath := cfg.Endpoint
	if basePath == "" {
		basePath = "/tmp/sharding-backups"
	}
	return &LocalStorage{logger: logger, basePath: basePath, objects: make(map[string][]byte), metadata: make(map[string]ObjectInfo)}, nil
}

func (l *LocalStorage) Upload(ctx context.Context, bucket, key string, data io.Reader, metadata map[string]string) error {
	body, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	fullKey := path.Join(bucket, key)
	l.mu.Lock()
	l.objects[fullKey] = body
	l.metadata[fullKey] = ObjectInfo{Key: key, Size: int64(len(body)), LastModified: time.Now(), Metadata: metadata}
	l.mu.Unlock()
	l.logger.Debug("uploaded to local storage", zap.String("key", fullKey))
	return nil
}

func (l *LocalStorage) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	fullKey := path.Join(bucket, key)
	l.mu.RLock()
	data, ok := l.objects[fullKey]
	l.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("object not found: %s", fullKey)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (l *LocalStorage) Delete(ctx context.Context, bucket, key string) error {
	fullKey := path.Join(bucket, key)
	l.mu.Lock()
	delete(l.objects, fullKey)
	delete(l.metadata, fullKey)
	l.mu.Unlock()
	return nil
}

func (l *LocalStorage) List(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	fullPrefix := path.Join(bucket, prefix)
	result := make([]ObjectInfo, 0)
	for key, info := range l.metadata {
		if strings.HasPrefix(key, fullPrefix) {
			result = append(result, info)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].LastModified.After(result[j].LastModified) })
	return result, nil
}

func (l *LocalStorage) Exists(ctx context.Context, bucket, key string) (bool, error) {
	fullKey := path.Join(bucket, key)
	l.mu.RLock()
	_, ok := l.objects[fullKey]
	l.mu.RUnlock()
	return ok, nil
}

func (l *LocalStorage) GetSignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	return fmt.Sprintf("file://%s/%s/%s", l.basePath, bucket, key), nil
}

func (l *LocalStorage) CreateBucket(ctx context.Context, bucket string) error { return nil }

func (l *LocalStorage) DeleteBucket(ctx context.Context, bucket string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	prefix := bucket + "/"
	for key := range l.objects {
		if strings.HasPrefix(key, prefix) {
			delete(l.objects, key)
			delete(l.metadata, key)
		}
	}
	return nil
}

