package config

import (
	"fmt"

	"github.com/jtenhave/not-just-noise/lib/config"
)

type Config struct {
	MySQL struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"db_name"`
	} `json:"mysql"`
	Worker struct {
		MaxWorkers    int `json:"max_workers"`
		MaxBatchSize  int `json:"max_batch_size"`
		NoJobDelay    int `json:"no_job_delay"`
		JobBufferSize int `json:"job_buffer_size"`
	} `json:"worker"`
}

// LoadConfig loads the config from the given path. Returns the config and the first error encountered.
func LoadConfig(path string) (Config, error) {
	var cgf Config
	err := config.LoadJsonFile(path, &cgf)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load mysql config: %w", err)
	}

	return cgf, nil
}
