package utils

import (
	"log"
	"sync"
	"time"

	"github.com/opsmon/server/common"
	"github.com/opsmon/server/model"
)

const maxBulkRetries = 3

type Consumer struct {
	queue       *Queue
	sender      *Sender
	retryDelay  int
	bulkSize    int
	flushTimer  int
	stopChannel chan struct{}
	buffer      []model.NormalizedLog
	bufferMu    sync.Mutex
}

func NewConsumer(queue *Queue, sender *Sender, retryDelay int) *Consumer {
	log.Println("Consumer initialized with buffering")
	return &Consumer{
		queue:       queue,
		sender:      sender,
		retryDelay:  retryDelay,
		bulkSize:    common.AppConfig.BulkSize,
		flushTimer:  common.AppConfig.FlushSeconds,
		stopChannel: make(chan struct{}),
		buffer:      make([]model.NormalizedLog, 0, common.AppConfig.BulkSize),
	}
}

func (c *Consumer) Start(workerID int) {
	log.Printf("Consumer worker %d started with BulkSize=%d, FlushTimer=%ds",
		workerID, c.bulkSize, c.flushTimer)

	ticker := time.NewTicker(time.Duration(c.flushTimer) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChannel:
			log.Printf("Consumer worker %d stopping, flushing remaining buffer", workerID)
			c.flushBuffer(workerID)
			return

		case <-ticker.C:
			log.Printf("Consumer worker %d: Timer triggered, flushing buffer", workerID)
			c.flushBuffer(workerID)

		default:
			batch, ok := c.tryDequeue()
			if !ok {
				log.Printf("Worker %d: Queue closed — exiting", workerID)
				return
			}
			if batch == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			c.addLogsToBuffer(batch, workerID)
		}
	}
}

func (c *Consumer) tryDequeue() (*model.WorkerBatch, bool) {
	select {
	case batch, ok := <-c.queue.channel:
		return batch, ok
	default:
		return nil, true
	}
}

func (c *Consumer) addLogsToBuffer(batch *model.WorkerBatch, workerID int) {
	c.bufferMu.Lock()
	defer c.bufferMu.Unlock()

	for _, log := range batch.Logs {
		c.buffer = append(c.buffer, log)
	}

	log.Printf("Consumer worker %d: Added %d logs to buffer (Buffer: %d/%d)",
		workerID, len(batch.Logs), len(c.buffer), c.bulkSize)

	if len(c.buffer) >= c.bulkSize {
		log.Printf("Consumer worker %d: Buffer size reached, flushing", workerID)
		c.doFlush(workerID)
	}
}

func (c *Consumer) flushBuffer(workerID int) {
	c.bufferMu.Lock()
	defer c.bufferMu.Unlock()
	c.doFlush(workerID)
}

func (c *Consumer) doFlush(workerID int) {
	if len(c.buffer) == 0 {
		return
	}

	logsToSend := make([]model.NormalizedLog, len(c.buffer))
	copy(logsToSend, c.buffer)

	log.Printf("Consumer worker %d: Flushing %d logs to Elasticsearch", workerID, len(logsToSend))

	for attempt := 1; attempt <= maxBulkRetries; attempt++ {

		err := c.sender.SendBulkToElasticsearch(logsToSend)
		if err == nil {
			log.Printf("Consumer worker %d: Successfully flushed %d logs", workerID, len(logsToSend))
			c.buffer = c.buffer[:0]
			return
		}

		log.Printf("Consumer worker %d: Bulk flush failed attempt %d/%d: %v",
			workerID, attempt, maxBulkRetries, err)

		time.Sleep(time.Duration(c.retryDelay) * time.Second)
	}

	log.Printf("Consumer worker %d: Dropping %d logs after %d retries",
		workerID, len(logsToSend), maxBulkRetries)

	c.buffer = c.buffer[:0]
}

func (c *Consumer) Stop() {
	close(c.stopChannel)
	log.Println("Consumer stop signal sent")
}
