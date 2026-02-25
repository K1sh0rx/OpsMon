package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/opsmon/server/model"
)

type Sender struct {
	elasticsearchURL string
	client           *http.Client
}

func NewSender(elasticsearchURL string) *Sender {
	log.Printf("Sender initialized with Elasticsearch URL: %s", elasticsearchURL)
	return &Sender{
		elasticsearchURL: elasticsearchURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendBulkToElasticsearch sends multiple logs in a single bulk request
func (s *Sender) SendBulkToElasticsearch(logs []model.NormalizedLog) error {
	if len(logs) == 0 {
		return nil
	}

	// Build Elasticsearch bulk request body
	var bulkBody bytes.Buffer
	for _, log := range logs {
		// Index action line
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": "logger",
			},
		}
		actionBytes, _ := json.Marshal(action)
		bulkBody.Write(actionBytes)
		bulkBody.WriteByte('\n')

		// Document line
		docBytes, _ := json.Marshal(log)
		bulkBody.Write(docBytes)
		bulkBody.WriteByte('\n')
	}

	// Send bulk request
	url := fmt.Sprintf("%s/_bulk", s.elasticsearchURL)
	req, err := http.NewRequest("POST", url, &bulkBody)
	if err != nil {
		return fmt.Errorf("failed to create bulk request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-ndjson")

	log.Printf("Sending bulk request to Elasticsearch: %d logs", len(logs))

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send bulk request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Parse response to check for errors
		var bulkResp map[string]interface{}
		if err := json.Unmarshal(body, &bulkResp); err == nil {
			if errors, ok := bulkResp["errors"].(bool); ok && errors {
				log.Printf("Warning: Bulk request had some errors: %s", string(body))
			} else {
				log.Printf("Successfully indexed %d logs to Elasticsearch", len(logs))
			}
		}
		return nil
	}

	return fmt.Errorf("elasticsearch returned error: status=%d, body=%s", resp.StatusCode, string(body))
}
