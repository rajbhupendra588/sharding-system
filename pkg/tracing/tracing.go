package tracing

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TracerConfig holds configuration for distributed tracing
type TracerConfig struct {
	ServiceName    string
	Environment    string
	JaegerEndpoint string
	ZipkinEndpoint string
	OTLPEndpoint   string
	SampleRate     float64
	Enabled        bool
}

// Tracer provides distributed tracing functionality
type Tracer struct {
	logger      *zap.Logger
	config      TracerConfig
	exporter    SpanExporter
	sampler     Sampler
	mu          sync.RWMutex
	activeSpans map[string]*Span
}

// SpanExporter defines interface for exporting spans
type SpanExporter interface {
	Export(ctx context.Context, spans []*Span) error
	Shutdown(ctx context.Context) error
}

// Sampler decides whether to sample a trace
type Sampler interface {
	ShouldSample(traceID string) bool
}

// Span represents a single operation within a trace
type Span struct {
	TraceID       string                 `json:"trace_id"`
	SpanID        string                 `json:"span_id"`
	ParentSpanID  string                 `json:"parent_span_id,omitempty"`
	OperationName string                 `json:"operation_name"`
	ServiceName   string                 `json:"service_name"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time,omitempty"`
	Duration      time.Duration          `json:"duration,omitempty"`
	Tags          map[string]string      `json:"tags,omitempty"`
	Logs          []SpanLog              `json:"logs,omitempty"`
	Status        SpanStatus             `json:"status"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// SpanLog represents a log entry within a span
type SpanLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
}

// SpanStatus represents the status of a span
type SpanStatus struct {
	Code        StatusCode `json:"code"`
	Description string     `json:"description,omitempty"`
}

// StatusCode represents span status codes
type StatusCode int

const (
	StatusCodeUnset StatusCode = iota
	StatusCodeOK
	StatusCodeError
)

// NewTracer creates a new tracer
func NewTracer(logger *zap.Logger, cfg TracerConfig) (*Tracer, error) {
	if !cfg.Enabled {
		return &Tracer{logger: logger, config: cfg, activeSpans: make(map[string]*Span)}, nil
	}

	var exporter SpanExporter
	var err error
	if cfg.JaegerEndpoint != "" {
		exporter, err = NewJaegerExporter(cfg.JaegerEndpoint)
	} else if cfg.ZipkinEndpoint != "" {
		exporter, err = NewZipkinExporter(cfg.ZipkinEndpoint)
	} else if cfg.OTLPEndpoint != "" {
		exporter, err = NewOTLPExporter(cfg.OTLPEndpoint)
	} else {
		exporter = &NoopExporter{}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	var sampler Sampler
	if cfg.SampleRate > 0 && cfg.SampleRate < 1 {
		sampler = &ProbabilitySampler{Rate: cfg.SampleRate}
	} else {
		sampler = &AlwaysSampler{}
	}

	return &Tracer{logger: logger, config: cfg, exporter: exporter, sampler: sampler, activeSpans: make(map[string]*Span)}, nil
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, operationName string, opts ...SpanOption) (context.Context, *Span) {
	var traceID, parentSpanID string
	if parentSpan := SpanFromContext(ctx); parentSpan != nil {
		traceID = parentSpan.TraceID
		parentSpanID = parentSpan.SpanID
	} else if traceIDCtx := ctx.Value(traceIDContextKey); traceIDCtx != nil {
		traceID = traceIDCtx.(string)
	} else {
		traceID = generateTraceID()
	}

	if !t.sampler.ShouldSample(traceID) {
		return ctx, nil
	}

	span := &Span{
		TraceID:       traceID,
		SpanID:        generateSpanID(),
		ParentSpanID:  parentSpanID,
		OperationName: operationName,
		ServiceName:   t.config.ServiceName,
		StartTime:     time.Now(),
		Tags:          make(map[string]string),
		Logs:          make([]SpanLog, 0),
		Attributes:    make(map[string]interface{}),
		Status:        SpanStatus{Code: StatusCodeUnset},
	}

	for _, opt := range opts {
		opt(span)
	}

	t.mu.Lock()
	t.activeSpans[span.SpanID] = span
	t.mu.Unlock()

	ctx = context.WithValue(ctx, spanContextKey, span)
	ctx = context.WithValue(ctx, traceIDContextKey, traceID)
	return ctx, span
}

// EndSpan ends a span and exports it
func (t *Tracer) EndSpan(span *Span) {
	if span == nil {
		return
	}
	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)

	t.mu.Lock()
	delete(t.activeSpans, span.SpanID)
	t.mu.Unlock()

	if t.exporter != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := t.exporter.Export(ctx, []*Span{span}); err != nil {
				t.logger.Warn("failed to export span", zap.Error(err))
			}
		}()
	}
}

// Shutdown shuts down the tracer
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.exporter != nil {
		return t.exporter.Shutdown(ctx)
	}
	return nil
}

// SpanOption is a function that configures a span
type SpanOption func(*Span)

// WithTag adds a tag to the span
func WithTag(key, value string) SpanOption {
	return func(s *Span) { s.Tags[key] = value }
}

// WithAttribute adds an attribute to the span
func WithAttribute(key string, value interface{}) SpanOption {
	return func(s *Span) { s.Attributes[key] = value }
}

// WithParentSpan sets the parent span
func WithParentSpan(parent *Span) SpanOption {
	return func(s *Span) {
		if parent != nil {
			s.TraceID = parent.TraceID
			s.ParentSpanID = parent.SpanID
		}
	}
}

func (s *Span) SetStatus(code StatusCode, description string) {
	if s == nil {
		return
	}
	s.Status = SpanStatus{Code: code, Description: description}
}

func (s *Span) SetTag(key, value string) {
	if s == nil {
		return
	}
	s.Tags[key] = value
}

func (s *Span) Log(fields map[string]interface{}) {
	if s == nil {
		return
	}
	s.Logs = append(s.Logs, SpanLog{Timestamp: time.Now(), Fields: fields})
}

func (s *Span) RecordError(err error) {
	if s == nil || err == nil {
		return
	}
	s.SetStatus(StatusCodeError, err.Error())
	s.Log(map[string]interface{}{"event": "error", "message": err.Error()})
}

type contextKey string

const (
	spanContextKey    contextKey = "span"
	traceIDContextKey contextKey = "trace_id"
)

// SpanFromContext extracts a span from context
func SpanFromContext(ctx context.Context) *Span {
	if span, ok := ctx.Value(spanContextKey).(*Span); ok {
		return span
	}
	return nil
}

// TraceIDFromContext extracts a trace ID from context
func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDContextKey).(string); ok {
		return traceID
	}
	return ""
}

// HTTPMiddleware creates HTTP middleware for tracing
func (t *Tracer) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = r.Header.Get("traceparent")
		}

		ctx := r.Context()
		if traceID != "" {
			ctx = context.WithValue(ctx, traceIDContextKey, traceID)
		}

		operationName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		ctx, span := t.StartSpan(ctx, operationName, WithTag("http.method", r.Method), WithTag("http.url", r.URL.String()), WithTag("http.user_agent", r.UserAgent()))

		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		if span != nil {
			w.Header().Set("X-Trace-ID", span.TraceID)
			w.Header().Set("X-Span-ID", span.SpanID)
		}

		next.ServeHTTP(wrapped, r.WithContext(ctx))

		if span != nil {
			span.SetTag("http.status_code", fmt.Sprintf("%d", wrapped.statusCode))
			if wrapped.statusCode >= 400 {
				span.SetStatus(StatusCodeError, fmt.Sprintf("HTTP %d", wrapped.statusCode))
			} else {
				span.SetStatus(StatusCodeOK, "")
			}
			t.EndSpan(span)
		}
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Samplers
type AlwaysSampler struct{}

func (s *AlwaysSampler) ShouldSample(traceID string) bool { return true }

type NeverSampler struct{}

func (s *NeverSampler) ShouldSample(traceID string) bool { return false }

type ProbabilitySampler struct{ Rate float64 }

func (s *ProbabilitySampler) ShouldSample(traceID string) bool { return rand.Float64() < s.Rate }

type RateLimitingSampler struct {
	MaxPerSecond int
	mu           sync.Mutex
	lastReset    time.Time
	count        int
}

func (s *RateLimitingSampler) ShouldSample(traceID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if now.Sub(s.lastReset) >= time.Second {
		s.count = 0
		s.lastReset = now
	}
	if s.count < s.MaxPerSecond {
		s.count++
		return true
	}
	return false
}

// Exporters
type NoopExporter struct{}

func (e *NoopExporter) Export(ctx context.Context, spans []*Span) error { return nil }
func (e *NoopExporter) Shutdown(ctx context.Context) error              { return nil }

type JaegerExporter struct {
	endpoint string
	client   *http.Client
}

func NewJaegerExporter(endpoint string) (*JaegerExporter, error) {
	return &JaegerExporter{endpoint: endpoint, client: &http.Client{Timeout: 5 * time.Second}}, nil
}

func (e *JaegerExporter) Export(ctx context.Context, spans []*Span) error { return nil }
func (e *JaegerExporter) Shutdown(ctx context.Context) error              { return nil }

type ZipkinExporter struct {
	endpoint string
	client   *http.Client
}

func NewZipkinExporter(endpoint string) (*ZipkinExporter, error) {
	return &ZipkinExporter{endpoint: endpoint, client: &http.Client{Timeout: 5 * time.Second}}, nil
}

func (e *ZipkinExporter) Export(ctx context.Context, spans []*Span) error { return nil }
func (e *ZipkinExporter) Shutdown(ctx context.Context) error              { return nil }

type OTLPExporter struct {
	endpoint string
	client   *http.Client
}

func NewOTLPExporter(endpoint string) (*OTLPExporter, error) {
	return &OTLPExporter{endpoint: endpoint, client: &http.Client{Timeout: 5 * time.Second}}, nil
}

func (e *OTLPExporter) Export(ctx context.Context, spans []*Span) error { return nil }
func (e *OTLPExporter) Shutdown(ctx context.Context) error              { return nil }

func generateTraceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateSpanID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// PropagateContext injects trace context into HTTP headers
func PropagateContext(ctx context.Context, req *http.Request) {
	span := SpanFromContext(ctx)
	if span != nil {
		req.Header.Set("X-Trace-ID", span.TraceID)
		req.Header.Set("X-Span-ID", span.SpanID)
		req.Header.Set("traceparent", fmt.Sprintf("00-%s-%s-01", span.TraceID, span.SpanID))
	} else if traceID := TraceIDFromContext(ctx); traceID != "" {
		req.Header.Set("X-Trace-ID", traceID)
	}
}

// ExtractContext extracts trace context from HTTP headers
func ExtractContext(r *http.Request) context.Context {
	ctx := r.Context()
	if traceparent := r.Header.Get("traceparent"); traceparent != "" {
		parts := splitTraceparent(traceparent)
		if len(parts) >= 3 {
			ctx = context.WithValue(ctx, traceIDContextKey, parts[1])
		}
	} else if traceID := r.Header.Get("X-Trace-ID"); traceID != "" {
		ctx = context.WithValue(ctx, traceIDContextKey, traceID)
	}
	return ctx
}

func splitTraceparent(s string) []string {
	result := make([]string, 0, 4)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '-' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

