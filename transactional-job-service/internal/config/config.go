package config

import (
	"fmt"

	"github.com/jtenhave/not-just-noise/lib/config"
	"github.com/jtenhave/not-just-noise/lib/database"
)

const mysqlConfigPath = "mysql.json"
const workerConfigPath = "worker.json"

type WorkerConfig struct {
	MaxWorkers    int `json:"max_workers"`
	MaxBatchSize  int `json:"max_batch_size"`
	NoJobDelay    int `json:"no_job_delay"`
	JobBufferSize int `json:"job_buffer_size"`
}

type Config struct {
	MySQL  database.MySQLConfig
	Worker WorkerConfig
}

// LoadConfig loads the config from the given path. Returns the config and the first error encountered.
func LoadConfig(path string) (Config, error) {
	var mySQLConfig database.MySQLConfig
	err := config.LoadJsonFile(fmt.Sprintf("%s%s", path, mysqlConfigPath), &mySQLConfig)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load mysql config: %w", err)
	}

	var workerConfig WorkerConfig
	err = config.LoadJsonFile(fmt.Sprintf("%s%s", path, workerConfigPath), &workerConfig)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load worker config: %w", err)
	}

	return Config{
		MySQL:  mySQLConfig,
		Worker: workerConfig,
	}, nil
}
