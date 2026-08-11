package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ActiveBot string `mapstructure:"active_bot"`
	Host      string `mapstructure:"host"`
	LogLevel  string `mapstructure:"log_level"`
	Port      uint16 `mapstructure:"port"`
}

func Load(env string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(env)
	v.AddConfigPath("config/")
	v.AutomaticEnv()

	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("port", 80)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config %q: %w", env, err)
	}

	var conf Config
	if err := v.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}
	return &conf, nil
}
