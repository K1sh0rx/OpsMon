package model

import "time"

// WorkerType identifies the log collection worker
type WorkerType string

const (
	JournaldWorker WorkerType = "journald"
	FileWorker     WorkerType = "file"
)

// NormalizedLog represents a single normalized log event
type NormalizedLog struct {
	Timestamp time.Time `json:"timestamp"`
	Host      string    `json:"host"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Facility  string    `json:"facility"`
	Transport string    `json:"transport"`

	// Process information (JournaldWorker)
	Process string `json:"process,omitempty"`
	PID     string `json:"pid,omitempty"`
	UID     string `json:"uid,omitempty"`
	GID     string `json:"gid,omitempty"`
	Exe     string `json:"exe,omitempty"`
	Cmdline string `json:"cmdline,omitempty"`

	// Audit information (JournaldWorker)
	AuditSession  string `json:"audit_session,omitempty"`
	AuditLoginUID string `json:"audit_loginuid,omitempty"`

	// Network information (FileWorker)
	SourceIP string `json:"source_ip,omitempty"`
}

// WorkerBatch represents a batch of logs from a single worker
type WorkerBatch struct {
	AgentID string          `json:"agent_id"`
	Worker  WorkerType      `json:"worker"`
	Logs    []NormalizedLog `json:"logs"`
}

// DeliveryResponse is sent back to the agent
type DeliveryResponse struct {
	Status      string `json:"status"`                // success | retry | error
	Description string `json:"description,omitempty"` // optional reason
}

// Response status constants
const (
	StatusSuccess = "success"
	StatusRetry   = "retry"
	StatusError   = "error"
)
