package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/opsmon/server/backend/engine"
)

type AlertsHandler struct {
	store *engine.AlertStore
}

func NewAlertsHandler() *AlertsHandler {
	return &AlertsHandler{
		store: engine.GetStore(),
	}
}

// GET /api/v1/alerts
func (h *AlertsHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	alerts, err := h.store.LoadAlerts()
	if err != nil {
		log.Printf("[alerts] Failed to load alerts: %v", err)
		http.Error(w, "Failed to load alerts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// PATCH /api/v1/alerts/{id}
func (h *AlertsHandler) UpdateAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/alerts/")
	id := strings.TrimSpace(path)

	if id == "" {
		http.Error(w, "Alert ID required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate allowed fields
	allowedFields := map[string]bool{"status": true, "description": true}
	for key := range updates {
		if !allowedFields[key] {
			http.Error(w, "Invalid field: "+key, http.StatusBadRequest)
			return
		}
	}

	// Validate status values if provided
	if status, ok := updates["status"].(string); ok {
		validStatuses := map[string]bool{
			"new":            true,
			"investigating":  true,
			"resolved":       true,
			"false_positive": true,
		}
		if !validStatuses[status] {
			http.Error(w, "Invalid status value", http.StatusBadRequest)
			return
		}
	}

	// Update alert
	if err := h.store.UpdateAlert(id, updates); err != nil {
		log.Printf("[alerts] Failed to update alert %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Alert updated",
	})
}

// DELETE /api/v1/alerts/{id}
func (h *AlertsHandler) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/alerts/")
	id := strings.TrimSpace(path)

	if id == "" {
		http.Error(w, "Alert ID required", http.StatusBadRequest)
		return
	}

	// Delete alert
	if err := h.store.DeleteAlert(id); err != nil {
		log.Printf("[alerts] Failed to delete alert %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Alert deleted",
	})
}

// HandleAlerts routes to appropriate handler based on path
func (h *AlertsHandler) HandleAlerts(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := r.URL.Path

	if path == "/api/v1/alerts" || path == "/api/v1/alerts/" {
		// GET /api/v1/alerts
		h.GetAlerts(w, r)
	} else if strings.HasPrefix(path, "/api/v1/alerts/") {
		// PATCH or DELETE /api/v1/alerts/{id}
		switch r.Method {
		case http.MethodPatch:
			h.UpdateAlert(w, r)
		case http.MethodDelete:
			h.DeleteAlert(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
