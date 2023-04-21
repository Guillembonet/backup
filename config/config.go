package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

var flagConfigPath = flag.String("configPath", "./example_config.yaml", "Path to config file")

type Config struct {
	Runtime Runtime `yaml:"runtime"`
	RunMode RunMode `yaml:"run_mode"`
	Backup  Backup  `yaml:"backup"`
}

type Runtime struct {
	LogLevel string `yaml:"log_level"`
}

type RunMode struct {
	RunOnceAndExit bool          `yaml:"run_once_and_exit"`
	Interval       time.Duration `yaml:"interval"`
}

type Backup struct {
	EncryptionPassword string   `yaml:"encryption_password"`
	Sources            []Source `yaml:"sources"`
	Targets            []Target `yaml:"targets"`
}

type Source struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

type Target struct {
	Type                 string            `yaml:"type"`
	BackupExpirationDays int               `yaml:"backup_expiration_days"`
	Config               map[string]string `yaml:"config"`
}

func LoadConfig() (*Config, error) {
	flag.Parsed()

	data, err := os.ReadFile(*flagConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}
	return &config, nil
}
