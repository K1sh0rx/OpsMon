package sender

import (
	"fmt"
	"log"
	"sync/atomic"
)

var fatalStop = make(chan struct{})

var fatalOnce atomic.Bool

const maxRetries = 2

// queueStopped is set to 1 atomically when SOC revokes the agent.
// All subsequent Send calls check this flag and bail immediately.
var queueStopped int32

// withRetry executes sendFn up to maxRetries times.
// Transport failures will retry.
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
			return lastErr
		}

		// Otherwise transport/network error → retry
		log.Printf("[retry] %s attempt %d/%d failed: %v",
			label,
			attempt,
			maxRetries,
			lastErr,
		)
	}

	// Transport failure only — DO NOT STOP AGENT
	log.Printf("[retry] %s transport failure after %d retries — will retry on next flush",
		label,
		maxRetries,
	)

	if fatalOnce.CompareAndSwap(false, true) {
		close(fatalStop)
	}

	return lastErr
}

func FatalStopChan() <-chan struct{} {
	return fatalStop
}
// IsStopped reports whether the queue has been stopped due to SOC revocation.
func IsStopped() bool {
	return atomic.LoadInt32(&queueStopped) == 1
}
