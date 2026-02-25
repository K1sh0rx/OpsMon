package utils

import (
	"log"

	"github.com/opsmon/server/model"
)

type Producer struct {
	queue *Queue
}

func NewProducer(queue *Queue) *Producer {
	log.Println("Producer initialized")
	return &Producer{
		queue: queue,
	}
}

func (p *Producer) Produce(batch *model.WorkerBatch) bool {
	if batch == nil {
		log.Println("Producer: Received nil batch, skipping")
		return false
	}

	if batch.AgentID == "" {
		log.Println("Producer: Invalid batch - missing AgentID")
		return false
	}

	if len(batch.Logs) == 0 {
		log.Printf("Producer: Empty log batch from AgentID=%s, skipping", batch.AgentID)
		return false
	}

	success := p.queue.Enqueue(batch)
	if !success {
		log.Printf("Producer: Failed to enqueue batch from AgentID=%s (queue full)", batch.AgentID)
	}
	return success
}
