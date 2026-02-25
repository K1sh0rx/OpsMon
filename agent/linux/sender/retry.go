package sender

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

var fatalStop = make(chan struct{})
var fatalOnce atomic.Bool

const maxRetries = 2

// queueStopped is set to 1 atomically when SOC revokes the agent.
// All subsequent Send calls check this flag and bail immediately.
var queueStopped int32

// ============================
// ✅ UPDATED RETRYABLE ERROR
// ============================
// Agent decides retry delay now

// withRetry executes sendFn up to maxRetries times.
// RetryableError will HOLD batch and retry forever after 30s.
// Unauthorized agent errors will STOP immediately.
func withRetry(label string, sendFn func() error) error {

	// If SOC already revoked agent
	if atomic.LoadInt32(&queueStopped) == 1 {
		return fmt.Errorf("queue stopped — dropping %s", label)
	}

	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {

		lastErr = sendFn()

		if lastErr == nil {
			return nil
		}

		// 🚨 STOP ONLY IF UNAUTHORIZED AGENT
		if _, ok := lastErr.(UnauthorizedError); ok {
			log.Printf("[retry] unauthorized agent — stopping queue")
			atomic.StoreInt32(&queueStopped, 1)

			if fatalOnce.CompareAndSwap(false, true) {
				close(fatalStop)
			}

			return lastErr
		}

		// ============================
		// 🟡 SOC BACKPRESSURE (retry)
		// ============================
		if _, ok := lastErr.(RetryableError); ok {

			log.Printf("[retry] SOC backpressure — holding batch and retrying every 30s")

			for {

				time.Sleep(30 * time.Second)

				if !checkServerReachable() {
					log.Printf("[retry] SOC still unreachable — waiting again")
					continue
				}

				log.Printf("[retry] SOC reachable — retrying same batch")

				err := sendFn()

				if err == nil {
					return nil
				}

				if _, ok := err.(UnauthorizedError); ok {
					log.Printf("[retry] agent revoked during retry — stopping queue")
					atomic.StoreInt32(&queueStopped, 1)

					if fatalOnce.CompareAndSwap(false, true) {
						close(fatalStop)
					}

					return err
				}

				log.Printf("[retry] retry failed — will retry again")
			}
		}

		// 🔁 Transport / Network Failure
		log.Printf("[retry] %s attempt %d/%d failed: %v",
			label,
			attempt,
			maxRetries,
			lastErr,
		)
	}

	// Transport failure exhausted — DO NOT STOP AGENT
	log.Printf("[retry] %s delivery failed after %d retries — will retry on next flush",
		label,
		maxRetries,
	)

	return lastErr
}

// FatalStopChan returns a channel that closes when SOC revokes the agent.
func FatalStopChan() <-chan struct{} {
	return fatalStop
}

// IsStopped reports whether the queue has been stopped due to SOC revocation.
func IsStopped() bool {
	return atomic.LoadInt32(&queueStopped) == 1
}
