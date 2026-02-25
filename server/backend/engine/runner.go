package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/opsmon/server/backend/engine/rules"
)

type RuleRunner struct {
	esClient *elasticsearch.Client
	store    *AlertStore
	firstRun bool
}

func NewRuleRunner(es *elasticsearch.Client) *RuleRunner {
	return &RuleRunner{
		esClient: es,
		store:    GetStore(),
		firstRun: true,
	}
}

func (r *RuleRunner) Start() {
	log.Println("[runner] Rule Engine Started")
	r.initializeCursor()

	go func() {
		for {
			r.runOnce()
			time.Sleep(10 * time.Second)
		}
	}()
}

// 🔥 OPEN PIT
func (r *RuleRunner) openPIT() (string, error) {

	res, err := r.esClient.OpenPointInTime(
		[]string{"logger"},
		"1m",
	)

	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var pitResp map[string]interface{}

	if err := json.NewDecoder(res.Body).Decode(&pitResp); err != nil {
		return "", err
	}

	id, ok := pitResp["id"].(string)
	if !ok {
		return "", err
	}

	return id, nil
}

func (r *RuleRunner) initializeCursor() {
	cursor, err := LoadCursor()
	if err != nil {
		log.Printf("[runner] Error loading cursor: %v", err)
		return
	}

	if cursor == nil {
		log.Println("[runner] No cursor found, initializing empty cursor")
		SaveCursor([]interface{}{})
	}
}

func (r *RuleRunner) runOnce() {

	searchAfter, _ := LoadCursor()

	// 🔥 OPEN PIT FIRST
	pitID, err := r.openPIT()
	if err != nil {
		log.Println("[runner] PIT open failed:", err)
		return
	}

	query := map[string]interface{}{
		"size":             100,
		"track_total_hits": false,
		"pit": map[string]interface{}{
			"id":         pitID,
			"keep_alive": "1m",
		},
		"sort": []interface{}{
			map[string]interface{}{
				"timestamp": map[string]interface{}{
					"order":         "asc",
					"unmapped_type": "date",
				},
			},
			map[string]interface{}{
				"_shard_doc": map[string]interface{}{
					"order": "asc",
				},
			},
		},
	}

	if searchAfter != nil && len(searchAfter) == 2 {
		query["search_after"] = searchAfter
	}

	data, _ := json.Marshal(query)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ❌ DO NOT SEND INDEX WITH PIT
	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		log.Println("[runner] ES error:", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[runner] Elasticsearch returned error: %s", res.String())
		return
	}

	var esResp map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		log.Printf("[runner] Failed to decode ES response: %v", err)
		return
	}

	hitsObj := esResp["hits"].(map[string]interface{})
	hits := hitsObj["hits"].([]interface{})

	if len(hits) == 0 {
		if r.firstRun {
			log.Println("[runner] No logs found on first run")
			r.firstRun = false
		}
		return
	}

	log.Printf("[runner] Processing %d logs", len(hits))

	for _, h := range hits {

		hit := h.(map[string]interface{})
		source := hit["_source"].(map[string]interface{})
		docID := hit["_id"].(string)

		r.runRules(source, docID)
	}

	last := hits[len(hits)-1].(map[string]interface{})
	sortVal := last["sort"].([]interface{})

	if len(sortVal) == 2 {
		SaveCursor(sortVal)
		log.Printf("[runner] Cursor saved: %v", sortVal)
	}
}

func (r *RuleRunner) runRules(logDoc map[string]interface{}, docID string) {

	// SSH BRUTE FORCE
	if rules.IsSSHBruteForce(logDoc) && !r.alertExists(docID,"ssh_bruteforce") {

		now := time.Now().UTC()

		alert := Alert{
			ID:          "ssh-" + now.Format("20060102150405.000000000"),
			Timestamp:   now,
			CreatedAt:   now,
			UpdatedAt:   now,
			Severity:    "high",
			Title:       "Failed SSH Login Attempt",
			Description: getStr(logDoc, "message"),
			RuleName:    "ssh_bruteforce",
			Host:        getStr(logDoc, "host"),
			LogID:       docID,
			Status:      "new",
		}

		log.Println("[ALERT CREATED] SSH Failed Login")
		r.store.AddAlert(alert)
	}

	// WEB COMMAND INJECTION
	if rules.IsWebCommandInjection(logDoc) && !r.alertExists(docID,"web_cmd_injection") {

		now := time.Now().UTC()

		alert := Alert{
			ID:          "web-" + now.Format("20060102150405.000000000"),
			Timestamp:   now,
			CreatedAt:   now,
			UpdatedAt:   now,
			Severity:    "critical",
			Title:       "Web Command Injection Attempt",
			Description: getStr(logDoc, "message"),
			RuleName:    "web_cmd_injection",
			Host:        getStr(logDoc, "host"),
			LogID:       docID,
			Status:      "new",
		}

		log.Println("[ALERT CREATED] Web Payload Detected")
		r.store.AddAlert(alert)
	}
}

func (r *RuleRunner) alertExists(logID string, rule string) bool {
	alerts, _ := r.store.LoadAlerts()

	for _, a := range alerts {
		if a.LogID == logID && a.RuleName == rule {
			return true
		}
	}
	return false
}

func getStr(m map[string]interface{}, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}
