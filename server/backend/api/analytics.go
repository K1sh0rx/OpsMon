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
	"github.com/opsmon/server/backend/engine"
)

type AnalyticsHandler struct {
	esClient *elasticsearch.Client
	store    *engine.AlertStore
}

func NewAnalyticsHandler(esClient *elasticsearch.Client) *AnalyticsHandler {
	return &AnalyticsHandler{
		esClient: esClient,
		store:    engine.GetStore(),
	}
}

type TimeSeriesPoint struct {
	Time  string `json:"time"`
	Count int64  `json:"count"`
}

func (h *AnalyticsHandler) GetIngestionTrend(w http.ResponseWriter, r *http.Request) {
	h.handleTrend(w, r, "all")
}

func (h *AnalyticsHandler) GetErrorTrend(w http.ResponseWriter, r *http.Request) {
	h.handleTrend(w, r, "error")
}

func (h *AnalyticsHandler) GetWarningTrend(w http.ResponseWriter, r *http.Request) {
	h.handleTrend(w, r, "warning")
}

func (h *AnalyticsHandler) handleTrend(w http.ResponseWriter, r *http.Request, severity string) {

	h.enableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "24h"
	}

	data, err := h.getTimeSeries(rangeParam, severity)
	if err != nil {
		log.Printf("[analytics] Failed: %v", err)
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *AnalyticsHandler) GetAlertTrend(w http.ResponseWriter, r *http.Request) {

	h.enableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "24h"
	}

	start, _ := getTimeRange(rangeParam)
	interval := getInterval(rangeParam)

	var points []TimeSeriesPoint
	current := start

	alerts, _ := h.store.LoadAlerts()

	for i := 0; i < 10; i++ {

		next := current.Add(interval)
		count := 0

		for _, alert := range alerts {
			if (alert.CreatedAt.Equal(current) || alert.CreatedAt.After(current)) &&
				alert.CreatedAt.Before(next) {
				count++
			}
		}

		points = append(points, TimeSeriesPoint{
			Time:  current.Format(time.RFC3339),
			Count: int64(count),
		})

		current = next
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(points)
}

func (h *AnalyticsHandler) getTimeSeries(rangeParam, severity string) ([]TimeSeriesPoint, error) {

	query := buildTimeSeriesQuery(rangeParam, severity)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.esClient.Search(
		h.esClient.Search.WithContext(ctx),
		h.esClient.Search.WithIndex("logger"),
		h.esClient.Search.WithBody(bytes.NewReader(query)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("ES error: %s", string(body))
	}

	return parseTimeSeriesResponse(res.Body)
}

func getESInterval(rangeParam string) (string, string) {

	switch rangeParam {
	case "24h":
		return "now-24h", "144m"
	case "all":
		return "now-30d", "72h"
	default:
		return "now-24h", "144m"
	}
}

func buildTimeSeriesQuery(rangeParam, severity string) []byte {

	var gte string
	var interval string

	switch rangeParam {
	case "24h":
		gte = "now-24h"
		interval = "144m"   // 24h / 10
	case "all":
		gte = "now-30d"
		interval = "72h"    // 30d / 10
	default:
		gte = "now-24h"
		interval = "144m"
	}

	queryMap := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": gte,
								"lte": "now",
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"time_series": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":          "timestamp",
					"fixed_interval": interval,
					"min_doc_count":  0,
					"extended_bounds": map[string]interface{}{
						"min": gte,
						"max": "now",
					},
				},
			},
		},
	}

	if severity != "all" {
		must := queryMap["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"severity.keyword": severity,
			},
		})
		queryMap["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = must
	}

	data, _ := json.Marshal(queryMap)
	return data
}

func parseTimeSeriesResponse(body io.Reader) ([]TimeSeriesPoint, error) {

	var esResp map[string]interface{}
	if err := json.NewDecoder(body).Decode(&esResp); err != nil {
		return nil, err
	}

	var points []TimeSeriesPoint

	if aggs, ok := esResp["aggregations"].(map[string]interface{}); ok {
		if ts, ok := aggs["time_series"].(map[string]interface{}); ok {
			if buckets, ok := ts["buckets"].([]interface{}); ok {

				for _, b := range buckets {

					bucket := b.(map[string]interface{})
					key := int64(bucket["key"].(float64))

					count := int64(0)
					if dc, ok := bucket["doc_count"].(float64); ok {
						count = int64(dc)
					}

					t := time.UnixMilli(key)

					points = append(points, TimeSeriesPoint{
						Time:  t.Format(time.RFC3339),
						Count: count,
					})
				}
			}
		}
	}

	return points, nil
}

func getTimeRange(rangeParam string) (time.Time, time.Time) {

	end := time.Now().UTC()

	switch rangeParam {
	case "24h":
		return end.Add(-24 * time.Hour), end
	case "all":
		return end.Add(-30 * 24 * time.Hour), end
	default:
		return end.Add(-24 * time.Hour), end
	}
}

func getInterval(rangeParam string) time.Duration {

	switch rangeParam {
	case "24h":
		return 144 * time.Minute
	case "all":
		return 72 * time.Hour
	default:
		return 144 * time.Minute
	}
}

func (h *AnalyticsHandler) enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
