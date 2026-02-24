package state

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/K1sh0rx/OpsMon/agent/linux/common"
)

const stateFile = "/var/lib/opsmon/agent_state.json"

// LoadState reads persisted AgentState from disk.
// Returns a fresh empty state (and an error) if the file doesn't exist yet.
func LoadState() (*common.AgentState, error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("[state] no state file found, initializing fresh state")
			return &common.AgentState{}, err
		}
		return &common.AgentState{}, err
	}

	var s common.AgentState
	if err := json.Unmarshal(data, &s); err != nil {
		log.Printf("[state] corrupt state file, starting fresh: %v", err)
		return &common.AgentState{}, err
	}

	log.Printf("[state] loaded state: agent_id=%s registered=%v cursor=%s offset=%d",
		s.AgentID, s.Registered, s.DeliveryState.JournalCursor, s.DeliveryState.FileOffset)
	return &s, nil
}

// SaveState writes AgentState to disk atomically (write-then-rename).
func SaveState(s *common.AgentState) error {
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Write to a temp file first, then rename — atomic on Linux
	tmp := stateFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, stateFile); err != nil {
		return err
	}

	log.Printf("[state] state saved: agent_id=%s registered=%v cursor=%s offset=%d",
		s.AgentID, s.Registered, s.DeliveryState.JournalCursor, s.DeliveryState.FileOffset)
	return nil
}
