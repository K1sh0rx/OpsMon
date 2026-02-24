package normalizer

import (
	"strings"
	"time"

	"github.com/K1sh0rx/OpsMon/agent/linux/common"
)

// JournaldEntry is the raw JSON structure from `journalctl -o json`.
// Only the fields we care about are mapped.
type JournaldEntry struct {
	Message          string `json:"MESSAGE"`
	Priority         string `json:"PRIORITY"`
	SyslogFacility   string `json:"SYSLOG_FACILITY"`
	SyslogIdentifier string `json:"SYSLOG_IDENTIFIER"`
	Hostname         string `json:"_HOSTNAME"`
	PID              string `json:"_PID"`
	UID              string `json:"_UID"`
	GID              string `json:"_GID"`
	Exe              string `json:"_EXE"`
	Cmdline          string `json:"_CMDLINE"`
	AuditSession     string `json:"_AUDIT_SESSION"`
	AuditLoginUID    string `json:"_AUDIT_LOGINUID"`
	// Realtime timestamp in microseconds since epoch (journald format)
	RealtimeTimestamp string `json:"__REALTIME_TIMESTAMP"`
	Cursor            string `json:"__CURSOR"`
}

// NormalizeJournald converts a raw JournaldEntry into a NormalizedLog.
func NormalizeJournald(e JournaldEntry) common.NormalizedLog {
	return common.NormalizedLog{
		Timestamp:     parseJournaldTimestamp(e.RealtimeTimestamp),
		Host:          e.Hostname,
		Message:       e.Message,
		Severity:      mapPriority(e.Priority),
		Facility:      mapFacility(e.SyslogFacility),
		Transport:     "journald",
		Process:       e.SyslogIdentifier,
		PID:           e.PID,
		UID:           e.UID,
		GID:           e.GID,
		Exe:           e.Exe,
		Cmdline:       trimCmdline(e.Cmdline),
		AuditSession:  e.AuditSession,
		AuditLoginUID: e.AuditLoginUID,
	}
}

// parseJournaldTimestamp converts journald's microsecond epoch string to time.Time.
func parseJournaldTimestamp(usec string) time.Time {
	if usec == "" {
		return time.Now().UTC()
	}
	var us int64
	for _, c := range usec {
		if c < '0' || c > '9' {
			break
		}
		us = us*10 + int64(c-'0')
	}
	return time.Unix(us/1_000_000, (us%1_000_000)*1000).UTC()
}

// mapPriority maps syslog priority number to human label.
func mapPriority(p string) string {
	switch strings.TrimSpace(p) {
	case "0":
		return "emergency"
	case "1":
		return "alert"
	case "2":
		return "critical"
	case "3":
		return "error"
	case "4":
		return "warning"
	case "5":
		return "notice"
	case "6":
		return "info"
	case "7":
		return "debug"
	default:
		return "unknown"
	}
}

// mapFacility maps syslog facility number to human label.
func mapFacility(f string) string {
	switch strings.TrimSpace(f) {
	case "0":
		return "kern"
	case "1":
		return "user"
	case "2":
		return "mail"
	case "3":
		return "daemon"
	case "4":
		return "auth"
	case "6":
		return "lpr"
	case "9":
		return "cron"
	case "10":
		return "authpriv"
	case "16":
		return "local0"
	case "17":
		return "local1"
	default:
		return "system"
	}
}

// trimCmdline trims and shortens extremely long command lines.
func trimCmdline(c string) string {
	c = strings.TrimSpace(c)
	if len(c) > 256 {
		return c[:256] + "..."
	}
	return c
}
