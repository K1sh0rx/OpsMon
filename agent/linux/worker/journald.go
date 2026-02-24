package worker

import (
	"context"
	"log"

	"github.com/K1sh0rx/OpsMon/agent/linux/buffer"
	"github.com/K1sh0rx/OpsMon/agent/linux/collector"
	"github.com/K1sh0rx/OpsMon/agent/linux/common"
	"github.com/K1sh0rx/OpsMon/agent/linux/normalizer"
	"github.com/K1sh0rx/OpsMon/agent/linux/sender"
)

// StartJournaldWorker reads from journald (resuming from saved cursor),
// normalizes each entry, and pushes LogEnvelopes into the queue.
// Stops when ctx is cancelled or the queue/sender is stopped.
func StartJournaldWorker(ctx context.Context, queue *buffer.Queue, agentState *common.AgentState) {
	log.Println("[worker/journald] starting")

	cursor := agentState.DeliveryState.JournalCursor
	stream := collector.StreamJournald(ctx, cursor)

	for {
		select {
		case <-ctx.Done():
			log.Println("[worker/journald] context cancelled, stopping")
			return

		case entry, ok := <-stream:
			if !ok {
				log.Println("[worker/journald] stream closed")
				return
			}

			if sender.IsStopped() {
				log.Println("[worker/journald] sender stopped, dropping entry")
				return
			}

			normalized := normalizer.NormalizeJournald(entry)
			envelope := common.LogEnvelope{
				Worker:        common.JournaldWorker,
				Log:           normalized,
				JournalCursor: entry.Cursor,
			}

			select {
			case queue.Ch <- envelope:
			case <-ctx.Done():
				return
			}
		}
	}
}
