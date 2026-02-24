package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/K1sh0rx/OpsMon/agent/linux/common"
)

const (
	serverURL   = "http://localhost:8080" // override via ENV: OPSMON_SERVER
	logsPath    = "/logs"
	httpTimeout = 10 * time.Second
)

var httpClient = &http.Client{Timeout: httpTimeout}

func baseURL() string {
	if u := os.Getenv("OPSMON_SERVER"); u != "" {
		return u
	}
	return serverURL
}

// ============================
// UNAUTHORIZED ERROR TYPE
// ============================

type UnauthorizedError struct {
	msg string
}

func (e UnauthorizedError) Error() string {
	return e.msg
}

// ============================
// SEND
// ============================

func Send(batch common.WorkerBatch, agentState *common.AgentState) error {

	if agentState.AgentID == "" {
		return fmt.Errorf("agent not provisioned")
	}

	label := fmt.Sprintf("batch[worker=%s logs=%d]", batch.Worker, len(batch.Logs))

	return withRetry(label, func() error {
		return postBatch(batch)
	})
}

// ============================
// POST BATCH
// ============================

func postBatch(batch common.WorkerBatch) error {

	body, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("marshal batch: %w", err)
	}

	resp, err := httpClient.Post(baseURL()+logsPath, "application/json", bytes.NewReader(body))
	if err != nil {
		// 🚨 Transport error → retry
		return fmt.Errorf("transport error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("POST /logs: status %d", resp.StatusCode)
	}

	var dr common.DeliveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		log.Printf("[sender] could not decode delivery response: %v", err)
		return nil
	}

	// 🚨 SOC REJECTION
	if dr.Status != "success" {

		log.Printf("[sender] SERVER REJECTED AGENT: %s", dr.Description)

		return UnauthorizedError{
			msg: dr.Description,
		}
	}

	return nil
}

// ============================
// OUTBOUND IP
// ============================

func outboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "unknown"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}
