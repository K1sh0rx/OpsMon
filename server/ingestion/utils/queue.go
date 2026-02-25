package utils

import (
	"log"
	"sync"

	"github.com/opsmon/server/model"
)

type Queue struct {
	channel chan *model.WorkerBatch
	mu      sync.Mutex
	size    int
}

func NewQueue(size int) *Queue {
	log.Printf("Initializing queue with size: %d", size)
	return &Queue{
		channel: make(chan *model.WorkerBatch, size),
		size:    size,
	}
}

func (q *Queue) Enqueue(batch *model.WorkerBatch) bool {
	select {
	case q.channel <- batch:
		log.Printf("Batch enqueued: AgentID=%s, Worker=%s, LogCount=%d", 
			batch.AgentID, batch.Worker, len(batch.Logs))
		return true
	default:
		log.Printf("Queue full! Unable to enqueue batch from AgentID=%s", batch.AgentID)
		return false
	}
}

func (q *Queue) Dequeue() (*model.WorkerBatch, bool) {
	batch, ok := <-q.channel
	return batch, ok
}

func (q *Queue) Size() int {
	return len(q.channel)
}

func (q *Queue) Close() {
	close(q.channel)
	log.Println("Queue closed")
}
