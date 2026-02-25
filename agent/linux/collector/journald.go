package collector

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os/exec"
)

// RAW journald structure (safe decode)
type JournaldRawEntry struct {
	Message            any    `json:"MESSAGE"`
	Priority           string `json:"PRIORITY"`
	SyslogFacility     string `json:"SYSLOG_FACILITY"`
	SyslogIdentifier   string `json:"SYSLOG_IDENTIFIER"`
	Hostname           string `json:"_HOSTNAME"`
	PID                string `json:"_PID"`
	UID                string `json:"_UID"`
	GID                string `json:"_GID"`
	Exe                string `json:"_EXE"`
	Cmdline            string `json:"_CMDLINE"`
	AuditSession       string `json:"_AUDIT_SESSION"`
	AuditLoginUID      string `json:"_AUDIT_LOGINUID"`
	RealtimeTimestamp  string `json:"__REALTIME_TIMESTAMP"`
	Cursor             string `json:"__CURSOR"`
}

func StreamJournald(ctx context.Context, cursor string) <-chan JournaldRawEntry {
	out := make(chan JournaldRawEntry, 64)

	go func() {
		defer close(out)

		var args []string

		if cursor == "" {
			log.Println("[collector/journald] first start — reading last 24h logs")
			args = []string{"-o", "json", "--since", "24 hours ago", "-f"}
		} else {
			log.Printf("[collector/journald] resuming from cursor: %s", cursor)
			args = []string{"-o", "json", "-f", "--after-cursor=" + cursor}
		}

		cmd := exec.CommandContext(ctx, "journalctl", args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("[collector/journald] stdout pipe error: %v", err)
			return
		}

		if err := cmd.Start(); err != nil {
			log.Printf("[collector/journald] failed to start journalctl: %v", err)
			return
		}

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 256*1024), 256*1024)

		for scanner.Scan() {
			var entry JournaldRawEntry
			if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
				log.Printf("[collector/journald] unmarshal error: %v", err)
				continue
			}

			select {
			case out <- entry:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}
