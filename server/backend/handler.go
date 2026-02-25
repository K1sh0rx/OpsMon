package backend

import (
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/opsmon/server/backend/api"
)

// BackendModule manages the monitoring backend (Module 2)
type BackendModule struct {
	esClient          *elasticsearch.Client
	dashboardHandler  *api.DashboardHandler
	analyticsHandler  *api.AnalyticsHandler
	alertsHandler     *api.AlertsHandler
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

	// Initialize handlers
	dashboardHandler := api.NewDashboardHandler(esClient)
	analyticsHandler := api.NewAnalyticsHandler(esClient)
	alertsHandler := api.NewAlertsHandler()

	return &BackendModule{
		esClient:         esClient,
		dashboardHandler: dashboardHandler,
		analyticsHandler: analyticsHandler,
		alertsHandler:    alertsHandler,
	}, nil
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
