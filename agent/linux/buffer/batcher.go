package buffer

import (
	"context"
	"log"
	"time"

	"github.com/K1sh0rx/OpsMon/agent/linux/common"
	"github.com/K1sh0rx/OpsMon/agent/linux/sender"
)



// StartBatchConsumer drains the queue, groups logs into per-worker batches of
// up to batchSize, and ships them via sender. On sender failure the retry
// logic inside sender will stop further sends after maxRetries.
//
// Called as a goroutine from main.
func StartBatchConsumer(ctx context.Context, queue *Queue, agentState *common.AgentState , batchSize int,flushTimeout time.Duration) {
	log.Println("[batcher] starting batch consumer")

	// Per-worker accumulators: worker type → pending envelopes
	pending := make(map[common.WorkerType][]common.LogEnvelope)
	ticker := time.NewTicker(flushTimeout)
	defer ticker.Stop()

	flush := func(worker common.WorkerType) {
		entries := pending[worker]
		if len(entries) == 0 {
			return
		}

		batch := buildBatch(worker, entries, agentState)
		log.Printf("[batcher] flushing %d logs for worker=%s", len(entries), worker)

		if err := sender.Send(batch, agentState); err != nil {
			log.Printf("[batcher] send failed for worker=%s: %v — queue stopped", worker, err)
			// sender.Send already stopped further retries; nothing more to do.
			return
		}

		// Advance checkpoint: keep the last envelope's cursor/offset
		last := entries[len(entries)-1]
		updateCheckpoint(agentState, worker, last)

		// Clear accumulator
		pending[worker] = pending[worker][:0]
	}

	flushAll := func() {
		for w := range pending {
			flush(w)
		}
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("[batcher] context cancelled — flushing remaining logs")
			// Drain whatever is still sitting in the channel
			draining := true
			for draining {
				select {
				case env := <-queue.Ch:
					pending[env.Worker] = append(pending[env.Worker], env)
				default:
					draining = false
				}
			}
			flushAll()
			return

		case env := <-queue.Ch:
			pending[env.Worker] = append(pending[env.Worker], env)
			if len(pending[env.Worker]) >= batchSize {
				flush(env.Worker)
			}

		case <-ticker.C:
			flushAll()
		}
	}
}

// buildBatch constructs a WorkerBatch from accumulated envelopes.
func buildBatch(worker common.WorkerType, entries []common.LogEnvelope, state *common.AgentState) common.WorkerBatch {
	logs := make([]common.NormalizedLog, len(entries))
	for i, e := range entries {
		logs[i] = e.Log
	}
	return common.WorkerBatch{
		AgentID:    state.AgentID,
		Worker:     worker,
		Logs:       logs,
	}
}

// updateCheckpoint saves the latest cursor/offset back into agentState
// so main can persist it on shutdown.
func updateCheckpoint(state *common.AgentState, worker common.WorkerType, last common.LogEnvelope) {
	switch worker {
	case common.JournaldWorker:
		if last.JournalCursor != "" {
			state.DeliveryState.JournalCursor = last.JournalCursor
		}
	case common.FileWorker:
		if last.FileOffset > 0 {
			state.DeliveryState.FileOffset = last.FileOffset
		}
	}
}
