package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jtenhave/not-just-noise/lib/database"
)

const mysqlConfigPath = "mysql.json"

type Config struct {
	MySQL database.MySQLConfig
}

// LoadConfig loads the config from the given path. Returns the config and the first error encountered.
func LoadConfig(path string) (Config, error) {
	var mySQLConfig database.MySQLConfig
	err := loadJsonFile(fmt.Sprintf("%s%s", path, mysqlConfigPath), &mySQLConfig)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load mysql config: %w", err)
	}

	return Config{
		MySQL: mySQLConfig,
	}, nil
}

// loadJsonFile loads the json file at the given path into the given output. Returns the first error encountered.
func loadJsonFile(path string, output interface{}) error {
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
