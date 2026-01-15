package world

import (
	"encoding/json"
	"os"
)

func LoadWorld(filename string) (*World, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	w := &World{}
	if err := json.Unmarshal(data, w); err != nil {
		return nil, err
	}

	return w, nil
}

func SaveWorld(w *World, filename string) error {
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
