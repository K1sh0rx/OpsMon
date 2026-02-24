package buffer

import (
	"github.com/K1sh0rx/OpsMon/agent/linux/common"
)

// Queue is a bounded in-memory channel buffer for LogEnvelope items.
// Producers (workers) call Enqueue; the batch consumer reads from Ch.
type Queue struct {
	Ch chan common.LogEnvelope
}

// NewQueue creates a Queue with the given capacity.
// When full, Enqueue blocks — this applies back-pressure to workers.
func NewQueue(capacity int) *Queue {
	return &Queue{
		Ch: make(chan common.LogEnvelope, capacity),
	}
}

// Enqueue adds a log envelope to the queue.
// Blocks if the queue is full (back-pressure).
func (q *Queue) Enqueue(entry common.LogEnvelope) {
	q.Ch <- entry
}

// Len returns the current number of items waiting in the queue.
func (q *Queue) Len() int {
	return len(q.Ch)
}
