package common

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort        string
	QueueSize         int
	RetryDelay        int
	ElasticsearchURL  string
	ConsumerWorkers   int
	BulkSize          int
	FlushSeconds      int
}

var AppConfig *Config

func LoadConfig() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	queueSize, err := strconv.Atoi(getEnv("QUEUE_SIZE", "1000"))
	if err != nil {
		log.Fatal("Invalid QUEUE_SIZE value")
	}

	retryDelay, err := strconv.Atoi(getEnv("RETRY_DELAY", "30"))
	if err != nil {
		log.Fatal("Invalid RETRY_DELAY value")
	}

	consumerWorkers, err := strconv.Atoi(getEnv("CONSUMER_WORKERS", "3"))
	if err != nil {
		log.Fatal("Invalid CONSUMER_WORKERS value")
	}

	bulkSize, err := strconv.Atoi(getEnv("BULK_SIZE", "100"))
	if err != nil {
		log.Fatal("Invalid BULK_SIZE value")
	}

	flushSeconds, err := strconv.Atoi(getEnv("FLUSH_SECONDS", "10"))
	if err != nil {
		log.Fatal("Invalid FLUSH_SECONDS value")
	}

	AppConfig = &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		QueueSize:         queueSize,
		RetryDelay:        retryDelay,
		ElasticsearchURL:  getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		ConsumerWorkers:   consumerWorkers,
		BulkSize:          bulkSize,
		FlushSeconds:      flushSeconds,
	}

	log.Printf("Configuration loaded: Port=%s, QueueSize=%d, RetryDelay=%ds, ElasticsearchURL=%s, ConsumerWorkers=%d, BulkSize=%d, FlushSeconds=%ds",
		AppConfig.ServerPort, AppConfig.QueueSize, AppConfig.RetryDelay, AppConfig.ElasticsearchURL, 
		AppConfig.ConsumerWorkers, AppConfig.BulkSize, AppConfig.FlushSeconds)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
