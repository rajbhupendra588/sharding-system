package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogFormat represents the log output format
type LogFormat string

const (
	LogFormatJSON    LogFormat = "json"
	LogFormatConsole LogFormat = "console"
)

// LogLevel represents logging severity level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level              LogLevel  `json:"level"`
	Format             LogFormat `json:"format"`
	OutputPaths        []string  `json:"output_paths"`
	EnableCaller       bool      `json:"enable_caller"`
	EnableStack        bool      `json:"enable_stack"`
	LokiEndpoint       string    `json:"loki_endpoint,omitempty"`
	SamplingInitial    int       `json:"sampling_initial"`
	SamplingThereafter int       `json:"sampling_thereafter"`
}

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
	config    LogConfig
	exporters []LogExporter
	mu        sync.RWMutex
}

// LogExporter defines interface for log exporters
type LogExporter interface {
	Export(entry *LogEntry) error
	Close() error
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Logger      string                 `json:"logger,omitempty"`
	Caller      string                 `json:"caller,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	ServiceName string                 `json:"service_name,omitempty"`
}

// NewLogger creates a new logger with the given configuration
func NewLogger(cfg LogConfig) (*Logger, error) {
	if cfg.Level == "" {
		cfg.Level = LogLevelInfo
	}
	if cfg.Format == "" {
		cfg.Format = LogFormatJSON
	}
	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}

	var level zapcore.Level
	switch cfg.Level {
	case LogLevelDebug:
		level = zapcore.DebugLevel
	case LogLevelInfo:
		level = zapcore.InfoLevel
	case LogLevelWarn:
		level = zapcore.WarnLevel
	case LogLevelError:
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	var encoderConfig zapcore.EncoderConfig
	if cfg.Format == LogFormatJSON {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       cfg.Format == LogFormatConsole,
		Encoding:          string(cfg.Format),
		EncoderConfig:     encoderConfig,
		OutputPaths:       cfg.OutputPaths,
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: !cfg.EnableStack,
		DisableCaller:     !cfg.EnableCaller,
	}

	if cfg.SamplingInitial > 0 {
		zapConfig.Sampling = &zap.SamplingConfig{Initial: cfg.SamplingInitial, Thereafter: cfg.SamplingThereafter}
	}

	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{Logger: zapLogger, config: cfg, exporters: make([]LogExporter, 0)}, nil
}

// AddExporter adds a log exporter
func (l *Logger) AddExporter(exporter LogExporter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.exporters = append(l.exporters, exporter)
}

// WithContext returns a logger with context fields
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	fields := make([]zap.Field, 0)
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		fields = append(fields, zap.String("trace_id", traceID.(string)))
	}
	if spanID := ctx.Value(SpanIDKey); spanID != nil {
		fields = append(fields, zap.String("span_id", spanID.(string)))
	}
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}
	return l.Logger.With(fields...)
}

// Close closes the logger and all exporters
func (l *Logger) Close() error {
	l.Logger.Sync()
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, exporter := range l.exporters {
		exporter.Close()
	}
	return nil
}

// Context keys
type contextKey string

const (
	TraceIDKey   contextKey = "trace_id"
	SpanIDKey    contextKey = "span_id"
	RequestIDKey contextKey = "request_id"
)

// LokiExporter exports logs to Grafana Loki
type LokiExporter struct {
	endpoint      string
	labels        map[string]string
	client        *http.Client
	batch         []*LogEntry
	batchSize     int
	mu            sync.Mutex
	stopCh        chan struct{}
	flushInterval time.Duration
}

// LokiExporterConfig holds Loki exporter configuration
type LokiExporterConfig struct {
	Endpoint      string
	Labels        map[string]string
	BatchSize     int
	FlushInterval time.Duration
}

// NewLokiExporter creates a new Loki exporter
func NewLokiExporter(cfg LokiExporterConfig) *LokiExporter {
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}
	if cfg.FlushInterval == 0 {
		cfg.FlushInterval = 5 * time.Second
	}

	exporter := &LokiExporter{
		endpoint:      cfg.Endpoint,
		labels:        cfg.Labels,
		client:        &http.Client{Timeout: 10 * time.Second},
		batch:         make([]*LogEntry, 0, cfg.BatchSize),
		batchSize:     cfg.BatchSize,
		stopCh:        make(chan struct{}),
		flushInterval: cfg.FlushInterval,
	}
	go exporter.backgroundFlusher()
	return exporter
}

func (e *LokiExporter) Export(entry *LogEntry) error {
	e.mu.Lock()
	e.batch = append(e.batch, entry)
	shouldFlush := len(e.batch) >= e.batchSize
	e.mu.Unlock()
	if shouldFlush {
		return e.flush()
	}
	return nil
}

func (e *LokiExporter) flush() error {
	e.mu.Lock()
	if len(e.batch) == 0 {
		e.mu.Unlock()
		return nil
	}
	batch := e.batch
	e.batch = make([]*LogEntry, 0, e.batchSize)
	e.mu.Unlock()

	streams := make([]lokiStream, 0)
	values := make([][]string, 0, len(batch))
	for _, entry := range batch {
		timestamp := fmt.Sprintf("%d", entry.Timestamp.UnixNano())
		line, _ := json.Marshal(entry)
		values = append(values, []string{timestamp, string(line)})
	}
	streams = append(streams, lokiStream{Stream: e.labels, Values: values})
	payload := lokiPushPayload{Streams: streams}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", e.endpoint+"/loki/api/v1/push", io.NopCloser(&jsonReader{data: body}))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("loki push failed with status %d", resp.StatusCode)
	}
	return nil
}

func (e *LokiExporter) backgroundFlusher() {
	ticker := time.NewTicker(e.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.flush()
		}
	}
}

func (e *LokiExporter) Close() error {
	close(e.stopCh)
	return e.flush()
}

type lokiPushPayload struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type jsonReader struct {
	data []byte
	pos  int
}

func (r *jsonReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// FileExporter exports logs to files with rotation
type FileExporter struct {
	path       string
	maxSize    int64
	maxBackups int
	file       *os.File
	size       int64
	mu         sync.Mutex
}

// FileExporterConfig holds file exporter configuration
type FileExporterConfig struct {
	Path       string
	MaxSize    int64
	MaxBackups int
}

// NewFileExporter creates a new file exporter
func NewFileExporter(cfg FileExporterConfig) (*FileExporter, error) {
	if cfg.MaxSize == 0 {
		cfg.MaxSize = 100 * 1024 * 1024
	}
	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = 5
	}

	exporter := &FileExporter{path: cfg.Path, maxSize: cfg.MaxSize, maxBackups: cfg.MaxBackups}
	if err := exporter.openFile(); err != nil {
		return nil, err
	}
	return exporter, nil
}

func (e *FileExporter) openFile() error {
	file, err := os.OpenFile(e.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}
	e.file = file
	e.size = info.Size()
	return nil
}

func (e *FileExporter) Export(entry *LogEntry) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	n, err := e.file.Write(data)
	if err != nil {
		return err
	}
	e.size += int64(n)

	if e.size >= e.maxSize {
		return e.rotate()
	}
	return nil
}

func (e *FileExporter) rotate() error {
	if e.file != nil {
		e.file.Close()
	}
	for i := e.maxBackups - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", e.path, i)
		newPath := fmt.Sprintf("%s.%d", e.path, i+1)
		os.Rename(oldPath, newPath)
	}
	os.Rename(e.path, e.path+".1")
	os.Remove(fmt.Sprintf("%s.%d", e.path, e.maxBackups+1))
	return e.openFile()
}

func (e *FileExporter) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.file != nil {
		return e.file.Close()
	}
	return nil
}

// RequestLogger middleware for HTTP request logging
type RequestLogger struct {
	logger *Logger
}

// NewRequestLogger creates a new request logger middleware
func NewRequestLogger(logger *Logger) *RequestLogger {
	return &RequestLogger{logger: logger}
}

// Middleware returns HTTP middleware for request logging
func (rl *RequestLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		if traceID := r.Header.Get("X-Trace-ID"); traceID != "" {
			ctx = context.WithValue(ctx, TraceIDKey, traceID)
		}
		if spanID := r.Header.Get("X-Span-ID"); spanID != "" {
			ctx = context.WithValue(ctx, SpanIDKey, spanID)
		}

		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		duration := time.Since(start)
		rl.logger.Info("http_request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", duration),
			zap.String("request_id", requestID),
			zap.String("user_agent", r.UserAgent()),
			zap.String("remote_addr", r.RemoteAddr),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
}

