package utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/opsmon/server/model"
)

type Receiver struct {
	producer *Producer
}

func NewReceiver(producer *Producer) *Receiver {
	log.Println("Receiver initialized")
	return &Receiver{
		producer: producer,
	}
}

func (r *Receiver) HandleLogs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		r.sendResponse(w, http.StatusMethodNotAllowed, model.StatusError, "Only POST method allowed")
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Receiver: Failed to read request body: %v", err)
		r.sendResponse(w, http.StatusBadRequest, model.StatusError, "Failed to read request body")
		return
	}
	defer req.Body.Close()

	var batch model.WorkerBatch
	if err := json.Unmarshal(body, &batch); err != nil {
		log.Printf("Receiver: Failed to unmarshal request: %v", err)
		r.sendResponse(w, http.StatusBadRequest, model.StatusError, "Invalid JSON payload")
		return
	}

	// Validate batch
	if err := r.validateBatch(&batch); err != nil {
		log.Printf("Receiver: Validation failed for AgentID=%s: %v", batch.AgentID, err)
		r.sendResponse(w, http.StatusBadRequest, model.StatusError, err.Error())
		return
	}

	log.Printf("Receiver: Received valid batch from AgentID=%s, Worker=%s, LogCount=%d",
		batch.AgentID, batch.Worker, len(batch.Logs))

	// Try to produce to queue
	success := r.producer.Produce(&batch)
	if !success {
		log.Printf("Receiver: Queue full, sending retry response to AgentID=%s", batch.AgentID)
		r.sendResponse(w, http.StatusServiceUnavailable, model.StatusRetry, "Queue is full, retry later")
		return
	}

	r.sendResponse(w, http.StatusOK, model.StatusSuccess, "Batch accepted")
}

func (r *Receiver) validateBatch(batch *model.WorkerBatch) error {
	if batch.AgentID == "" {
		return &ValidationError{Message: "agent_id is required"}
	}

	if batch.Worker != model.JournaldWorker && batch.Worker != model.FileWorker {
		return &ValidationError{Message: "invalid worker type"}
	}

	if len(batch.Logs) == 0 {
		return &ValidationError{Message: "logs array cannot be empty"}
	}

	// Validate each log
	for i, logEntry := range batch.Logs {
		if logEntry.Timestamp.IsZero() {
			return &ValidationError{Message: "log timestamp is required"}
		}
		if logEntry.Host == "" {
			return &ValidationError{Message: "log host is required"}
		}
		if logEntry.Message == "" {
			return &ValidationError{Message: "log message is required"}
		}
		if i >= 100 { // Prevent excessive validation on large batches
			break
		}
	}

	return nil
}

func (r *Receiver) sendResponse(w http.ResponseWriter, statusCode int, status, description string) {
	response := model.DeliveryResponse{
		Status:      status,
		Description: description,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)

	log.Printf("Receiver: Sent response - Status=%s, Code=%d, Description=%s",
		status, statusCode, description)
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
