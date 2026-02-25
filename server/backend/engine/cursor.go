// cursor.go
package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ✅ FIX: Use working directory + relative path for portability
var cursorFile string

func init() {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		// Fallback to relative path
		cursorFile = "backend/engine/cursor.json"
	} else {
		cursorFile = filepath.Join(wd, "backend", "engine", "cursor.json")
	}
}

type Cursor struct {
	SearchAfter []interface{} `json:"search_after"`
}

func LoadCursor() ([]interface{}, error) {
	data, err := os.ReadFile(cursorFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var c Cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return c.SearchAfter, nil
}

func SaveCursor(sa []interface{}) error {
	// ✅ FIX: Create directory if not exists
	dir := filepath.Dir(cursorFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	c := Cursor{
		SearchAfter: sa,
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cursorFile, data, 0644)
}
