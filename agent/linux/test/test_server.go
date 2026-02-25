package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ============================
// STORED BATCH STRUCT
// ============================

type StoredBatch struct {
	ReceivedAt time.Time   `json:"received_at"`
	Worker     string      `json:"worker"`
	BatchSize  int         `json:"batch_size"`
	Data       interface{} `json:"data"`
}

// ============================
// AGENT WHITELIST STRUCT
// ============================

type Agent struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
}

var (
	fileMutex sync.Mutex
	storeFile = "batches.json"
)

// ============================
// WHITELIST CHECK
// ============================

func isAgentAllowed(agentID string) bool {

	data, err := os.ReadFile("agents.json")
	if err != nil {
		log.Println("[server] agents.json missing")
		return false
	}

	var agents []Agent
	err = json.Unmarshal(data, &agents)
	if err != nil {
		return false
	}

	for _, a := range agents {
		if a.AgentID == agentID && a.Status == "active" {
			return true
		}
	}

	return false
}

// ============================
// LOGS HANDLER
// ============================

func logs(w http.ResponseWriter, r *http.Request) {

	var batch map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	agentID, ok := batch["agent_id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(map[string]string{
			"status":      "error",
			"description": "missing agent_id",
		})
		return
	}

	//  WHITELIST CHECK
	if !isAgentAllowed(agentID) {

		log.Printf("[server] rejected agent=%s", agentID)

		json.NewEncoder(w).Encode(map[string]string{
			"status":      "error",
			"description": "unauthorized agent",
		})
		return
	}

	logs, _ := batch["logs"].([]interface{})
	worker, _ := batch["worker"].(string)

	stored := StoredBatch{
		ReceivedAt: time.Now().UTC(),
		Worker:     worker,
		BatchSize:  len(logs),
		Data:       batch,
	}

	saveBatch(stored)

	resp := map[string]string{
		"status": "success",
	}
	json.NewEncoder(w).Encode(resp)
}

// ============================
// SAVE TO FILE
// ============================

func saveBatch(b StoredBatch) {

	fileMutex.Lock()
	defer fileMutex.Unlock()

	f, err := os.OpenFile(storeFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644)
	if err != nil {
		log.Println("file open error:", err)
		return
	}
	defer f.Close()

	data, _ := json.Marshal(b)
	f.Write(data)
	f.Write([]byte("\n"))

	log.Printf("[server] stored %s batch size=%d",
		b.Worker,
		b.BatchSize,
	)
}

// ============================
// MAIN
// ============================

func main() {

	http.HandleFunc("/api/v1/logs", logs)

	log.Println("🚀 Whitelist Test Server Running on :8080")
	http.ListenAndServe(":8080", nil)
}
