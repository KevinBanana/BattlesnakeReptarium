package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type LogConfig struct {
	LogLevel string `yaml:"level"`
}

type Config struct {
	LogConfig   LogConfig `yaml:"logConfig"`
	Environment string
}

func (c *Config) GetConfig() error {
	configFile, err := os.ReadFile(fmt.Sprintf("../config/%s.yml", c.Environment))
	if err != nil {
		return errors.Wrap(err, "GetConfig::os.ReadFile")
	}
	err = yaml.Unmarshal(configFile, c)
	if err != nil {
		log.Fatalf("getConfig::Unmarshal: %v", err)
	}
	return nil
}
