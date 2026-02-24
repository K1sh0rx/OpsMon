package worker

import (
	"context"
	"log"
	"os"

	"github.com/K1sh0rx/OpsMon/agent/linux/buffer"
	"github.com/K1sh0rx/OpsMon/agent/linux/collector"
	"github.com/K1sh0rx/OpsMon/agent/linux/common"
	"github.com/K1sh0rx/OpsMon/agent/linux/normalizer"
	"github.com/K1sh0rx/OpsMon/agent/linux/sender"
)

// StartNginxWorker tails the nginx access log (resuming from saved offset),
// normalizes each line, and pushes LogEnvelopes into the queue.
// Stops when ctx is cancelled or the sender is stopped.
func StartNginxWorker(ctx context.Context, queue *buffer.Queue, agentState *common.AgentState) {
	log.Println("[worker/nginx] starting")

	hostname, _ := os.Hostname()
	offset := agentState.DeliveryState.FileOffset
	stream := collector.StreamNginx(ctx, offset)

	for {
		select {
		case <-ctx.Done():
			log.Println("[worker/nginx] context cancelled, stopping")
			return

		case line, ok := <-stream:
			if !ok {
				log.Println("[worker/nginx] stream closed")
				return
			}

			if sender.IsStopped() {
				log.Println("[worker/nginx] sender stopped, dropping line")
				return
			}

			normalized, err := normalizer.NormalizeNginx(line.Text, hostname)
			if err != nil {
				// Silently skip malformed lines (e.g. partial writes at rotation)
				continue
			}

			envelope := common.LogEnvelope{
				Worker:     common.FileWorker,
				Log:        normalized,
				FileOffset: line.Offset,
			}

			select {
			case queue.Ch <- envelope:
			case <-ctx.Done():
				return
			}
		}
	}
}
