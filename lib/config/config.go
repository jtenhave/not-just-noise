package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// loadJsonFile loads the json file at the given path into the given output. Returns the first error encountered.
func LoadJsonFile(path string, output interface{}) error {
	rawFile, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	expandedFile := os.ExpandEnv(string(rawFile))

	err = json.Unmarshal([]byte(expandedFile), output)
	if err != nil {
		return fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return nil
}
