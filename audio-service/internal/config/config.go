package config

import (
	"fmt"

	"github.com/jtenhave/not-just-noise/lib/config"
	"github.com/jtenhave/not-just-noise/lib/database"
)

const mysqlConfigPath = "mysql.json"
const servicesConfigPath = "services.json"
const snsConfigPath = "sns.json"

type SNSConfig struct {
	TopicArn string `json:"topic_arn"`
}

type Config struct {
	MySQL database.MySQLConfig
	SNS   SNSConfig
}

// LoadConfig loads the config from the given path. Returns the config and the first error encountered.
func LoadConfig(path string) (Config, error) {
	var mySQLConfig database.MySQLConfig
	err := config.LoadJsonFile(fmt.Sprintf("%s%s", path, mysqlConfigPath), &mySQLConfig)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load mysql config: %w", err)
	}

	var snsConfig SNSConfig
	err = config.LoadJsonFile(fmt.Sprintf("%s%s", path, snsConfigPath), &snsConfig)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load sns config: %w", err)
	}

	return Config{
		MySQL: mySQLConfig,
		SNS:   snsConfig,
	}, nil
}
