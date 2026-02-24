package common

import "time"

//
// ==============================
// NORMALIZED EVENT (SIEM LOG)
// ==============================
// This is the final normalized log format emitted by all workers
//

type NormalizedLog struct {
	Timestamp       time.Time `json:"timestamp"`
	Host            string    `json:"host"`
	Message         string    `json:"message"`
	Severity        string    `json:"severity"`
	Facility        string    `json:"facility"`
	Transport       string    `json:"transport"`

	Process         string    `json:"process,omitempty"`
	PID             string    `json:"pid,omitempty"`
	UID             string    `json:"uid,omitempty"`
	GID             string    `json:"gid,omitempty"`
	Exe             string    `json:"exe,omitempty"`
	Cmdline         string    `json:"cmdline,omitempty"`

	AuditSession    string    `json:"audit_session,omitempty"`
	AuditLoginUID   string    `json:"audit_loginuid,omitempty"`

	SourceIP        string    `json:"source_ip,omitempty"` // file based logs (nginx)
}


type LogEnvelope struct {
    Worker WorkerType     `json:"worker"`
    Log    NormalizedLog  `json:"log"`

    // checkpoint info
    JournalCursor string `json:"journal_cursor,omitempty"`
    FileOffset    int64  `json:"file_offset,omitempty"`
}

//
// ==============================
// WORKER TYPES ENUM
// =============================
// Defines which collector produced the log
//

type WorkerType string

const (
	JournaldWorker WorkerType = "journald"
	FileWorker     WorkerType = "file"
)

//
// ==============================
// WORKER BATCH (UPLOAD UNIT)
// ==============================
// Each worker emits this independently
//

type WorkerBatch struct {
	AgentID    string          `json:"agent_id"`
	Worker     WorkerType      `json:"worker"`
	Logs       []NormalizedLog `json:"logs"`
}

//
// ==============================
// SERVER DELIVERY RESPONSE (ACK)
// ==============================
// Used by sender to confirm ingestion
//

type DeliveryResponse struct {
	Status      string `json:"status"`                // success | error
	Description string `json:"description,omitempty"` // error reason
}


//
// ==============================
// AGENT STATE (CRASH SAFE)
// ==============================
// Persist locally for resume logic
//

type DeliveryState struct {
	JournalCursor string `json:"journal_cursor"` // journald resume
	FileOffset    int64  `json:"file_offset"`    // nginx resume
}

type AgentState struct {
	AgentID     string      `json:"agent_id"`
	Registered  bool        `json:"registered"`
	DeliveryState DeliveryState `json:"worker_state"`
}


