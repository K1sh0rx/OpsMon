// store.go

package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ✅ FIX: Use working directory + relative path for portability
var alertsFile string

func init() {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		// Fallback to relative path
		alertsFile = "backend/engine/alerts.json"
	} else {
		alertsFile = filepath.Join(wd, "backend", "engine", "alerts.json")
	}
}

// Alert represents a security alert
type Alert struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"`    // critical, high, medium, low
	Title       string    `json:"title"`
	Description string    `json:"description"`
	RuleName    string    `json:"rule_name"`
	Host        string    `json:"host"`
	SourceIP    string    `json:"source_ip,omitempty"`
	LogID       string    `json:"log_id,omitempty"` // ES document _id
	Status      string    `json:"status"`           // new, investigating, resolved, false_positive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AlertStore provides thread-safe operations on alerts.json
type AlertStore struct {
	mu     sync.RWMutex
	alerts []Alert
}

var store *AlertStore
var once sync.Once

// GetStore returns the singleton AlertStore instance
func GetStore() *AlertStore {
	once.Do(func() {
		store = &AlertStore{
			alerts: []Alert{},
		}
		// Load existing alerts on initialization
		if err := store.load(); err != nil {
			// If file doesn't exist, start with empty array
			if !os.IsNotExist(err) {
				fmt.Printf("[store] Error loading alerts: %v\n", err)
			}
		}
	})
	return store
}

// LoadAlerts returns all alerts (read-only copy)
func (s *AlertStore) LoadAlerts() ([]Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	alertsCopy := make([]Alert, len(s.alerts))
	copy(alertsCopy, s.alerts)
	return alertsCopy, nil
}

// AddAlert appends a new alert and persists to disk
func (s *AlertStore) AddAlert(alert Alert) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set timestamps
	now := time.Now().UTC()
	alert.CreatedAt = now
	alert.UpdatedAt = now
	alert.Status = "new"

	s.alerts = append(s.alerts, alert)
	return s.save()
}

// UpdateAlert modifies an existing alert by ID
func (s *AlertStore) UpdateAlert(id string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.alerts {
		if s.alerts[i].ID == id {
			// Update allowed fields
			if status, ok := updates["status"].(string); ok {
				s.alerts[i].Status = status
			}
			if description, ok := updates["description"].(string); ok {
				s.alerts[i].Description = description
			}
			s.alerts[i].UpdatedAt = time.Now().UTC()
			return s.save()
		}
	}
	return fmt.Errorf("alert not found: %s", id)
}

// DeleteAlert removes an alert by ID
func (s *AlertStore) DeleteAlert(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.alerts {
		if s.alerts[i].ID == id {
			// Remove by slicing
			s.alerts = append(s.alerts[:i], s.alerts[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("alert not found: %s", id)
}

// GetAlertsByTimeRange filters alerts by time range
func (s *AlertStore) GetAlertsByTimeRange(start, end time.Time) ([]Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []Alert
	for _, alert := range s.alerts {
		if (alert.CreatedAt.Equal(start) || alert.CreatedAt.After(start)) &&
			(alert.CreatedAt.Equal(end) || alert.CreatedAt.Before(end)) {
			filtered = append(filtered, alert)
		}
	}
	return filtered, nil
}

// CountAlertsByTimeRange returns count of alerts in time range
func (s *AlertStore) CountAlertsByTimeRange(start, end time.Time) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, alert := range s.alerts {
		if (alert.CreatedAt.Equal(start) || alert.CreatedAt.After(start)) &&
			(alert.CreatedAt.Equal(end) || alert.CreatedAt.Before(end)) {
			count++
		}
	}
	return count
}

// load reads alerts from disk (caller must hold write lock)
func (s *AlertStore) load() error {
	data, err := os.ReadFile(alertsFile)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		s.alerts = []Alert{}
		return nil
	}

	if err := json.Unmarshal(data, &s.alerts); err != nil {
		return fmt.Errorf("unmarshal alerts: %w", err)
	}

	return nil
}

// save writes alerts to disk atomically (caller must hold write lock)
func (s *AlertStore) save() error {
	// Create directory if not exists
	dir := filepath.Dir(alertsFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	data, err := json.MarshalIndent(s.alerts, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal alerts: %w", err)
	}

	// Write to temp file then rename (atomic)
	tmpFile := alertsFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, alertsFile); err != nil {
		return fmt.Errorf("rename file: %w", err)
	}

	return nil
}
