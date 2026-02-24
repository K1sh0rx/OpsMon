package collector

import (
	"bufio"
	"context"
	"log"
	"os"
	"time"
)

const (
	nginxLogPath  = "/var/log/nginx/access.log"
	nginxPollInterval = 500 * time.Millisecond
)

// NginxLine carries a raw log line together with the file offset
// *after* the line — so the worker can checkpoint it.
type NginxLine struct {
	Text   string
	Offset int64
}

// StreamNginx tails the nginx access log starting at byteOffset,
// polling every 500ms for new lines. Handles log rotation via reopen.
// Streams NginxLine values to the returned channel.
func StreamNginx(ctx context.Context, byteOffset int64) <-chan NginxLine {
	out := make(chan NginxLine, 64)

	go func() {
		defer close(out)

		f, err := openAt(nginxLogPath, byteOffset)
		if err != nil {
			log.Printf("[collector/nginx] cannot open %s: %v", nginxLogPath, err)
			return
		}
		defer f.Close()

		currentOffset := byteOffset
		reader := bufio.NewReader(f)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				currentOffset += int64(len(line))
				select {
				case out <- NginxLine{Text: line, Offset: currentOffset}:
				case <-ctx.Done():
					return
				}
				continue
			}

			if err != nil {
				// EOF — check for rotation then poll
				if rotated(f) {
					log.Println("[collector/nginx] log rotation detected, reopening")
					f.Close()
					f, err = openAt(nginxLogPath, 0)
					if err != nil {
						log.Printf("[collector/nginx] reopen failed: %v", err)
						return
					}
					reader.Reset(f)
					currentOffset = 0
				}
				select {
				case <-time.After(nginxPollInterval):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}

// openAt opens the file and seeks to offset.
func openAt(path string, offset int64) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if offset > 0 {
		if _, err := f.Seek(offset, 0); err != nil {
			f.Close()
			return nil, err
		}
	}
	return f, nil
}

// rotated returns true if the open file descriptor no longer matches
// the inode on disk — indicating the file was rotated.
func rotated(f *os.File) bool {
	diskStat, err := os.Stat(nginxLogPath)
	if err != nil {
		return false
	}
	fdStat, err := f.Stat()
	if err != nil {
		return false
	}
	return !os.SameFile(diskStat, fdStat)
}
