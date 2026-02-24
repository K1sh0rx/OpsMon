package main

import (
	"context"
	"log"
	"os"
	"fmt"
	"os/signal"
	"syscall"
	"time"
	"strings"
	"github.com/joho/godotenv"


	"github.com/K1sh0rx/OpsMon/agent/linux/buffer"
	"github.com/K1sh0rx/OpsMon/agent/linux/state"
    "github.com/K1sh0rx/OpsMon/agent/linux/worker"
	"github.com/K1sh0rx/OpsMon/agent/linux/sender"
	"github.com/K1sh0rx/OpsMon/agent/linux/common"
)

func main() {

	godotenv.Load()

	log.Println("🚀 OpsMon Linux Agent Starting...")

	// -----------------------------
	// LOAD DELIVERY STATE
	// -----------------------------

	agentState, err := state.LoadState()
	if err != nil {
		log.Println("No previous state found. Starting fresh...")
	}

	if agentState.AgentID == "" {
	    log.Println("Agent not provisioned")
	    var id string
	    log.Print("Enter Agent ID provided by SOC: ")
	    fmt.Scanln(&id)
		id = strings.TrimSpace(id)
		if id == "" {
    		log.Fatal("Invalid Agent ID")
		}
	    agentState.AgentID = id
	    agentState.Registered = true
	    err := state.SaveState(agentState)
	    if err != nil {
	        log.Fatal("Failed saving AgentID:", err)
	    }
	    log.Println("Agent provisioned successfully")
	}

	// -----------------------------
	// GLOBAL CONTEXT
	// -----------------------------

	ctx, cancel := context.WithCancel(context.Background())

	// -----------------------------
	// QUEUE INIT (BOUNDED BUFFER)
	// -----------------------------

	queueSize := common.GetEnvInt("QUEUE_SIZE",1000)
	queue := buffer.NewQueue(queueSize)

	// -----------------------------
	// START WORKERS
	// -----------------------------

	go worker.StartJournaldWorker(ctx, queue, agentState)
	go worker.StartNginxWorker(ctx, queue, agentState)

	// -----------------------------
	// START BATCH CONSUMER
	// -----------------------------

	batchSize := common.GetEnvInt("BATCH_SIZE", 500)
    flushSec  := common.GetEnvInt("FLUSH_TIMEOUT", 5)

    flushTimeout := time.Duration(flushSec) * time.Second
	go buffer.StartBatchConsumer(ctx, queue, agentState, batchSize, flushTimeout)

	log.Println("Workers started... Waiting for SIGTERM")

	// -----------------------------
	// SIGNAL HANDLER
	// -----------------------------

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {

	case <-sigChan:
		log.Println("Shutdown signal received")

	case <-sender.FatalStopChan():
		log.Println("Fatal send failure — stopping agent")
    }


	// -----------------------------
	// STOP WORKERS
	// -----------------------------

	cancel()

	log.Println("Waiting for drain...")

	time.Sleep(5 * time.Second)

	// -----------------------------
	// FLUSH & SAVE STATE
	// -----------------------------

	err = state.SaveState(agentState)
	if err != nil {
		log.Println("Failed saving state:", err)
	}

	log.Println("Agent stopped safely ✔")
}
