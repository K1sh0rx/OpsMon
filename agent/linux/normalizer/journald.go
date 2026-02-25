package normalizer

import (
	"regexp"
	"strings"
	"time"

	"github.com/K1sh0rx/OpsMon/agent/linux/common"
	"github.com/K1sh0rx/OpsMon/agent/linux/collector"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func NormalizeJournald(e collector.JournaldRawEntry) common.NormalizedLog {

	msg := stringifyMessage(e.Message)

	if msg == "" {
		msg = strings.TrimSpace(e.SyslogIdentifier)
	}
	if msg == "" {
		msg = strings.TrimSpace(e.Exe)
	}
	if msg == "" {
		msg = trimCmdline(e.Cmdline)
	}
	if msg == "" {
		msg = "[JOURNALD_EMPTY]"
	}

	host := strings.TrimSpace(e.Hostname)
	if host == "" {
		host = "unknown"
	}

	return common.NormalizedLog{
		Timestamp:     parseJournaldTimestamp(e.RealtimeTimestamp),
		Host:          host,
		Message:       msg,
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

func stringifyMessage(msg any) string {

	switch v := msg.(type) {

	case string:
		return clean(v)

	case []any:
		b := make([]byte, len(v))
		for i, n := range v {
			if f, ok := n.(float64); ok {
				b[i] = byte(f)
			}
		}
		return clean(string(b))

	default:
		return ""
	}
}

func clean(s string) string {
	s = ansiRegex.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
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
