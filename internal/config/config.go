package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	ActiveBot string `mapstructure:"active_bot"`
	Host      string `mapstructure:"host"`
	LogLevel  string `mapstructure:"log_level"`
	Port      uint16 `mapstructure:"port"`
}

var parsedConfig Config

func Init(env string) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(env)
	v.AddConfigPath("../../config/")
	v.AddConfigPath("config/")
	v.AutomaticEnv()

	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("port", 8080)

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("error parsing configuration file, %v", err)
	}
	if err := v.Unmarshal(&parsedConfig); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}

func GetConfig() *Config {
	return &parsedConfig
}
