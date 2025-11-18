package security

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AuditLogger logs audit events
type AuditLogger struct {
	file   *os.File
	mu     sync.Mutex
	writer *json.Encoder
}

// AuditEvent represents an audit event
type AuditEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	User       string    `json:"user"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id,omitempty"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
	IP         string    `json:"ip,omitempty"`
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logPath string) (*AuditLogger, error) {
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	return &AuditLogger{
		file:   file,
		writer: json.NewEncoder(file),
	}, nil
}

// Log logs an audit event
func (a *AuditLogger) Log(event AuditEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	event.Timestamp = time.Now()
	if err := a.writer.Encode(event); err != nil {
		// Log error but don't fail the operation
		fmt.Fprintf(os.Stderr, "failed to write audit log: %v\n", err)
	}
}

// Close closes the audit logger
func (a *AuditLogger) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.file.Close()
}

