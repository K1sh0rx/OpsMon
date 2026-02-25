package ingestion

import (
	"fmt"
	"log"
	"sync"

	"github.com/opsmon/server/common"
	"github.com/opsmon/server/ingestion/utils"
)

type IngestionModule struct {
	queue     *utils.Queue
	producer  *utils.Producer
	consumers []*utils.Consumer
	sender    *utils.Sender
	receiver  *utils.Receiver
	wg        sync.WaitGroup
}

func NewIngestionModule() *IngestionModule {

	log.Println("=== Initializing Module 1: Ingestion ===")

	// Queue
	queue := utils.NewQueue(common.AppConfig.QueueSize)

	// Producer
	producer := utils.NewProducer(queue)

	// Sender → Elasticsearch
	sender := utils.NewSender(common.AppConfig.ElasticsearchURL)

	// Consumers
	consumers := make([]*utils.Consumer, common.AppConfig.ConsumerWorkers)
	for i := 0; i < common.AppConfig.ConsumerWorkers; i++ {
		consumers[i] = utils.NewConsumer(
			queue,
			sender,
			common.AppConfig.RetryDelay,
		)
	}

	// Receiver (HTTP handler)
	receiver := utils.NewReceiver(producer)

	log.Println("=== Module 1: Ingestion initialized successfully ===")

	return &IngestionModule{
		queue:     queue,
		producer:  producer,
		consumers: consumers,
		sender:    sender,
		receiver:  receiver,
	}
}

func (im *IngestionModule) Start() error {

	log.Println("=== Starting Module 1: Ingestion ===")

	for i, consumer := range im.consumers {
		im.wg.Add(1)
		go func(workerID int, c *utils.Consumer) {
			defer im.wg.Done()
			c.Start(workerID)
		}(i, consumer)
	}

	log.Printf("Started %d consumer workers", len(im.consumers))
	log.Println("=== Module 1: Ingestion started successfully ===")

	return nil
}

func (im *IngestionModule) Stop() error {

	log.Println("=== Stopping Module 1: Ingestion ===")

	for _, consumer := range im.consumers {
		consumer.Stop()
	}

	im.queue.Close()
	im.wg.Wait()

	log.Println("=== Module 1: Ingestion stopped ===")

	return nil
}

func (im *IngestionModule) GetReceiver() *utils.Receiver {
	return im.receiver
}

func (im *IngestionModule) GetStatus() string {
	return fmt.Sprintf(
		"Queue size: %d/%d, Consumers: %d",
		im.queue.Size(),
		common.AppConfig.QueueSize,
		len(im.consumers),
	)
}
