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
	S3 struct {
		TempBucket string `json:"temp_bucket"`
	} `json:"s3"`
	SQS struct {
		UploadQueueURL       string `json:"upload_queue_url"`
		UploadCommitQueueURL string `json:"upload_commit_queue_url"`
	} `json:"sqs"`
	Worker struct {
		MaxWorkers    int `json:"max_workers"`
		MaxBatchSize  int `json:"max_batch_size"`
		NoJobDelay    int `json:"no_job_delay"`
		JobBufferSize int `json:"job_buffer_size"`
	} `json:"worker"`
}

// LoadConfig loads the config from the given path. Returns the config and the first error encountered.
func LoadConfig(path string) (Config, error) {
	var cfg Config
	err := config.LoadJsonFile(path, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}
