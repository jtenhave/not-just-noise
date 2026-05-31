package config

import (
	"fmt"

	"github.com/jtenhave/not-just-noise/lib/config"
)

const configPath = "config.json"

type Config struct {
	MySQL struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"db_name"`
	} `json:"mysql"`
	SNS struct {
		TopicArn string `json:"topic_arn"`
	} `json:"sns"`
}

// LoadConfig loads the config from the given path. Returns the config and the first error encountered.
func LoadConfig(path string) (Config, error) {
	var cgf Config
	err := config.LoadJsonFile(fmt.Sprintf("%s%s", path, configPath), &cgf)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load mysql config: %w", err)
	}

	return cgf, nil
}
