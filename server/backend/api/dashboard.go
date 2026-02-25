package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type DashboardHandler struct {
	esClient *elasticsearch.Client
}

func NewDashboardHandler(esClient *elasticsearch.Client) *DashboardHandler {
	return &DashboardHandler{esClient: esClient}
}

type DashboardMetrics struct {
	TotalLogs      int64       `json:"total_logs"`
	LogsPerSec     float64     `json:"logs_per_sec"`
	ErrorCount     int64       `json:"error_count"`
	WarningCount   int64       `json:"warning_count"`
	TopHosts       []CountItem `json:"top_hosts"`
	TopProcesses   []CountItem `json:"top_processes"`
	TransportSplit []CountItem `json:"transport_split"`
	TimeRange      string      `json:"time_range"`
}

type CountItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

func (h *DashboardHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "24h"
	}

	query := buildDashboardQuery(rangeParam)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.esClient.Search(
		h.esClient.Search.WithContext(ctx),
		h.esClient.Search.WithIndex("logger"),
		h.esClient.Search.WithBody(bytes.NewReader(query)),
	)
	if err != nil {
		http.Error(w, "ES query failed", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		log.Printf("[dashboard] ES error: %s", string(body))
		http.Error(w, "ES error", http.StatusInternalServerError)
		return
	}

	metrics, err := parseDashboardResponse(res.Body, rangeParam)
	if err != nil {
		http.Error(w, "Parse error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func buildDashboardQuery(rangeParam string) []byte {

	timeFilter := "now-24h"
	if rangeParam == "all" {
		timeFilter = "now-30d"
	}

	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": timeFilter,
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{

			"total_logs": map[string]interface{}{
				"value_count": map[string]interface{}{
					"field": "timestamp",
				},
			},

			"error_count": map[string]interface{}{
				"filter": map[string]interface{}{
					"term": map[string]interface{}{
						"severity.keyword": "error",
					},
				},
			},

			"warning_count": map[string]interface{}{
				"filter": map[string]interface{}{
					"term": map[string]interface{}{
						"severity.keyword": "warning",
					},
				},
			},

			"top_hosts": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "host.keyword",
					"size":  10,
				},
			},

			"top_processes": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "process.keyword",
					"size":  10,
				},
			},

			"transport_split": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "transport.keyword",
					"size":  10,
				},
			},
		},
	}

	data, _ := json.Marshal(query)
	return data
}

func parseDashboardResponse(body io.Reader, rangeParam string) (*DashboardMetrics, error) {

	var esResp map[string]interface{}
	if err := json.NewDecoder(body).Decode(&esResp); err != nil {
		return nil, err
	}

	metrics := &DashboardMetrics{TimeRange: rangeParam}

	if aggs, ok := esResp["aggregations"].(map[string]interface{}); ok {

		if t, ok := aggs["total_logs"].(map[string]interface{}); ok {
			if v, ok := t["value"].(float64); ok {
				metrics.TotalLogs = int64(v)
			}
		}

		if e, ok := aggs["error_count"].(map[string]interface{}); ok {
			if v, ok := e["doc_count"].(float64); ok {
				metrics.ErrorCount = int64(v)
			}
		}

		if w, ok := aggs["warning_count"].(map[string]interface{}); ok {
			if v, ok := w["doc_count"].(float64); ok {
				metrics.WarningCount = int64(v)
			}
		}

		metrics.TopHosts = extractBuckets(aggs, "top_hosts")
		metrics.TopProcesses = extractBuckets(aggs, "top_processes")
		metrics.TransportSplit = extractBuckets(aggs, "transport_split")
	}

	duration := 24.0
	if rangeParam == "all" {
		duration = 24.0 * 30
	}

	if metrics.TotalLogs > 0 {
		metrics.LogsPerSec = float64(metrics.TotalLogs) / (duration * 3600)
	}

	return metrics, nil
}

func extractBuckets(aggs map[string]interface{}, name string) []CountItem {

	var items []CountItem

	if agg, ok := aggs[name].(map[string]interface{}); ok {
		if buckets, ok := agg["buckets"].([]interface{}); ok {

			for _, b := range buckets {

				bucket := b.(map[string]interface{})
				key := fmt.Sprintf("%v", bucket["key"])

				count := int64(0)
				if dc, ok := bucket["doc_count"].(float64); ok {
					count = int64(dc)
				}

				items = append(items, CountItem{
					Name:  key,
					Count: count,
				})
			}
		}
	}

	return items
}
