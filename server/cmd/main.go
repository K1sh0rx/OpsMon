package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opsmon/server/backend"
	"github.com/opsmon/server/common"
	"github.com/opsmon/server/ingestion"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("=================================================")
	log.Println("       OpsMon Server - Central Orchestrator     ")
	log.Println("=================================================")

	// Load ENV
	common.LoadConfig()

	// ---------------- MODULE INIT ----------------
	log.Println("\n--- Initializing Modules ---")

	ingestionModule := ingestion.NewIngestionModule()

	backendModule, err := backend.NewBackendModule(
		common.AppConfig.ElasticsearchURL,
	)
	if err != nil {
		log.Fatalf("Backend init failed: %v", err)
	}

	// ---------------- ROUTER ----------------
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/api/v1/logs",
		ingestionModule.GetReceiver().HandleLogs,
	)

	mux.HandleFunc("/health", healthCheck)

	backendModule.RegisterRoutes(mux)

	backendModule.StartEngine()

	// ---------------- HTTP SERVER ----------------
	server := &http.Server{
		Addr:         ":" + common.AppConfig.ServerPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ---------------- START INGESTION ----------------
	log.Println("\n--- Starting Modules ---")

	if err := ingestionModule.Start(); err != nil {
		log.Fatalf("Ingestion start failed: %v", err)
	}

	// ---------------- START SERVER ----------------
	go func() {

		log.Printf(
			"HTTP Server listening on :%s",
			common.AppConfig.ServerPort,
		)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	log.Println("\n=================================================")
	log.Println("       OpsMon Server is running                  ")
	log.Println("=================================================")

	log.Printf("Module 1 (Ingestion): %s\n",
		ingestionModule.GetStatus())

	log.Printf("Module 2 (Backend): %s\n",
		backendModule.GetStatus())

	log.Println("Press Ctrl+C to shutdown gracefully")

	// ---------------- WAIT CTRL+C ----------------
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("\nShutdown signal received")

	// ---------------- GRACEFUL SHUTDOWN ----------------
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	log.Println("Stopping ingestion module...")
	ingestionModule.Stop()

	log.Println("Stopping HTTP server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
}


func healthCheck(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(`
	{
		"status":"healthy",
		"modules":["ingestion","backend"]
	}
	`))
}
