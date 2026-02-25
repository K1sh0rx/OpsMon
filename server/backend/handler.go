package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/opsmon/server/backend/api"
	"github.com/opsmon/server/backend/engine"
)

// BackendModule manages the monitoring backend (Module 2)
type BackendModule struct {
	esClient          *elasticsearch.Client
	dashboardHandler  *api.DashboardHandler
	analyticsHandler  *api.AnalyticsHandler
	alertsHandler     *api.AlertsHandler
	ruleRunner        *engine.RuleRunner
}

// NewBackendModule initializes the backend module
func NewBackendModule(esURL string) (*BackendModule, error) {
	log.Println("=== Initializing Backend Module ===")

	// Initialize Elasticsearch client
	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
	}
	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Test connection
	res, err := esClient.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Elasticsearch connection failed: %s", res.String())
		return nil, err
	}

	log.Println("Elasticsearch connection established")

	// ✅ FIX: Create index with proper mappings
	if err := createLoggerIndex(esClient); err != nil {
		log.Printf("Warning: Failed to create logger index: %v", err)
		// Don't fail - index might already exist
	}

	// Initialize handlers
	dashboardHandler := api.NewDashboardHandler(esClient)
	analyticsHandler := api.NewAnalyticsHandler(esClient)
	alertsHandler := api.NewAlertsHandler()
	ruleRunner := engine.NewRuleRunner(esClient)

	return &BackendModule{
		esClient:         esClient,
		dashboardHandler: dashboardHandler,
		analyticsHandler: analyticsHandler,
		alertsHandler:    alertsHandler,
		ruleRunner:       ruleRunner,
	}, nil
}

// ✅ NEW: createLoggerIndex creates the logger index with proper mappings
func createLoggerIndex(esClient *elasticsearch.Client) error {
	indexName := "logger"

	// Check if index exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := esClient.Indices.Exists([]string{indexName},
		esClient.Indices.Exists.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Index already exists
	if res.StatusCode == 200 {
		log.Println("Elasticsearch index 'logger' already exists")
		return nil
	}

	// Create index with mappings
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"type": "date",
				},
				"host": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"message": map[string]interface{}{
					"type": "text",
				},
				"severity": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"facility": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"transport": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"process": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"pid": map[string]interface{}{
					"type": "keyword",
				},
				"uid": map[string]interface{}{
					"type": "keyword",
				},
				"gid": map[string]interface{}{
					"type": "keyword",
				},
				"exe": map[string]interface{}{
					"type": "text",
				},
				"cmdline": map[string]interface{}{
					"type": "text",
				},
				"audit_session": map[string]interface{}{
					"type": "keyword",
				},
				"audit_loginuid": map[string]interface{}{
					"type": "keyword",
				},
				"source_ip": map[string]interface{}{
					"type": "ip",
				},
			},
		},
	}

	mappingJSON, _ := json.Marshal(mapping)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	res2, err := esClient.Indices.Create(
		indexName,
		esClient.Indices.Create.WithBody(bytes.NewReader(mappingJSON)),
		esClient.Indices.Create.WithContext(ctx2),
	)
	if err != nil {
		return err
	}
	defer res2.Body.Close()

	if res2.IsError() {
		log.Printf("Failed to create index: %s", res2.String())
		return err
	}

	log.Println("Elasticsearch index 'logger' created successfully with mappings")
	return nil
}

// RegisterRoutes registers all backend API routes
func (bm *BackendModule) RegisterRoutes(mux *http.ServeMux) {
	log.Println("=== Registering Backend Routes ===")

	// Dashboard routes
	mux.HandleFunc("/api/v1/dashboard/metrics", bm.dashboardHandler.GetMetrics)

	// Analytics routes
	mux.HandleFunc("/api/v1/analytics/ingestion", bm.analyticsHandler.GetIngestionTrend)
	mux.HandleFunc("/api/v1/analytics/errors", bm.analyticsHandler.GetErrorTrend)
	mux.HandleFunc("/api/v1/analytics/warnings", bm.analyticsHandler.GetWarningTrend)
	mux.HandleFunc("/api/v1/analytics/alerts", bm.analyticsHandler.GetAlertTrend)

	// Alerts routes (handles /api/v1/alerts and /api/v1/alerts/{id})
	mux.HandleFunc("/api/v1/alerts", bm.alertsHandler.HandleAlerts)
	mux.HandleFunc("/api/v1/alerts/", bm.alertsHandler.HandleAlerts)

	log.Println("Backend routes registered:")
	log.Println("  GET  /api/v1/dashboard/metrics")
	log.Println("  GET  /api/v1/analytics/ingestion")
	log.Println("  GET  /api/v1/analytics/errors")
	log.Println("  GET  /api/v1/analytics/warnings")
	log.Println("  GET  /api/v1/analytics/alerts")
	log.Println("  GET  /api/v1/alerts")
	log.Println("  PATCH /api/v1/alerts/{id}")
	log.Println("  DELETE /api/v1/alerts/{id}")
}

// GetStatus returns the module status
func (bm *BackendModule) GetStatus() string {
	return "Backend module active"
}

func (bm *BackendModule) StartEngine() {
	log.Println("Starting Runtime Detection Engine...")
	bm.ruleRunner.Start()
}
