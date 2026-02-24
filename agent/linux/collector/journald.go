package collector

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os/exec"

	"github.com/K1sh0rx/OpsMon/agent/linux/normalizer"
)

// JournaldEntry mirrors normalizer.JournaldEntry — we decode into this then normalize.
// Re-exported here so the worker doesn't import normalizer directly.
type JournaldRawEntry = normalizer.JournaldEntry

// StreamJournald launches `journalctl -o json -f [--after-cursor=<cursor>]`
// and streams decoded JournaldRawEntry values to the returned channel.
// The channel is closed when ctx is cancelled or journalctl exits.
func StreamJournald(ctx context.Context, cursor string) <-chan JournaldRawEntry {
	out := make(chan JournaldRawEntry, 64)

	go func() {
		defer close(out)

		args := []string{"-o", "json", "-f", "-n", "0"}
		if cursor != "" {
			args = append(args, "--after-cursor="+cursor)
			log.Printf("[collector/journald] resuming from cursor: %s", cursor)
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
		// Journal entries can be large; increase buffer.
		scanner.Buffer(make([]byte, 256*1024), 256*1024)

		for scanner.Scan() {
			line := scanner.Bytes()
			var entry JournaldRawEntry
			if err := json.Unmarshal(line, &entry); err != nil {
				log.Printf("[collector/journald] unmarshal error: %v", err)
				continue
			}
			select {
			case out <- entry:
			case <-ctx.Done():
				return
			}
		}

		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			log.Printf("[collector/journald] scanner error: %v", err)
		}

		cmd.Wait()
	}()

	return out
}
