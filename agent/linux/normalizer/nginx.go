package normalizer

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/K1sh0rx/OpsMon/agent/linux/common"
)

// nginxCombinedRe matches nginx "combined" log format:
// $remote_addr - $remote_user [$time_local] "$request" $status $bytes "$http_referer" "$http_user_agent"
var nginxCombinedRe = regexp.MustCompile(
	`^(\S+)\s+-\s+(\S+)\s+\[([^\]]+)\]\s+"([^"]+)"\s+(\d+)\s+(\d+)\s+"([^"]*)"\s+"([^"]*)"`,
)

// NginxLogFormat is the time format nginx uses in combined logs.
const nginxTimeFormat = "02/Jan/2006:15:04:05 -0700"

// NormalizeNginx parses a single nginx access log line into a NormalizedLog.
// Returns an error if the line doesn't match the expected format.
func NormalizeNginx(line string, hostname string) (common.NormalizedLog, error) {
	line = strings.TrimSpace(line)
	m := nginxCombinedRe.FindStringSubmatch(line)
	if m == nil {
		return common.NormalizedLog{}, fmt.Errorf("line does not match nginx combined format")
	}

	remoteAddr := m[1]
	// m[2] = remote_user (usually "-")
	timeLocal := m[3]
	request := m[4]
	status := m[5]
	// m[6] = bytes sent
	// m[7] = referer
	// m[8] = user agent

	ts, err := time.Parse(nginxTimeFormat, timeLocal)
	if err != nil {
		ts = time.Now().UTC()
	}

	return common.NormalizedLog{
		Timestamp: ts.UTC(),
		Host:      hostname,
		Message:   fmt.Sprintf("%s %s", status, request),
		Severity:  statusToSeverity(status),
		Facility:  "daemon",
		Transport: "file",
		Process:   "nginx",
		SourceIP:  remoteAddr,
	}, nil
}

// statusToSeverity maps HTTP status codes to syslog severity labels.
func statusToSeverity(status string) string {
	if len(status) == 0 {
		return "unknown"
	}
	switch status[0] {
	case '5':
		return "error"
	case '4':
		return "warning"
	case '3':
		return "notice"
	default:
		return "info"
	}
}
