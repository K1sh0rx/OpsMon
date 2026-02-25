package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url" // ✅ ADDED
	"os"
	"time"

	"github.com/K1sh0rx/OpsMon/agent/linux/common"
)

const (
	serverURL   = "http://localhost:8080" // override via ENV: OPSMON_SERVER
	logsPath    = "/api/v1/logs"
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
// ✅ ADDED HOST EXTRACTOR
// ============================

func baseURLHost() string {

	u, err := url.Parse(baseURL())
	if err != nil {
		return "localhost"
	}

	return u.Hostname()
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
// ✅ ADDED RETRYABLE ERROR TYPE
// ============================

type RetryableError struct{}

func (e RetryableError) Error() string {
	return "server requested retry"
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

	// ============================
	// ✅ UPDATED SOC RESPONSE HANDLING
	// ============================

	switch dr.Status {

	case "success":
		return nil

	case "error":
		log.Printf("[sender] SERVER REVOKED AGENT: %s", dr.Description)
		return UnauthorizedError{
			msg: dr.Description,
		}

	case "retry":
		log.Printf("[sender] SERVER BACKPRESSURE — retry later")
		return RetryableError{}

	default:
		log.Printf("[sender] unknown SOC status: %s", dr.Status)
		return fmt.Errorf("unknown SOC status")
	}
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
