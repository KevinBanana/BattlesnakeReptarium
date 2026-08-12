package main

import (
	"log/slog"
	"os"

	"BattlesnakeReptarium/internal/config"
	"BattlesnakeReptarium/internal/logging"
	"BattlesnakeReptarium/internal/server"
)

func getEnv() string {
	environment := os.Getenv("ENVIRONMENT")

	if environment == "" {
		// default to local
		return "local"
	}
	return environment
}

func main() {
	env := getEnv()

	conf, err := config.Load(env)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	logging.Init(conf.LogLevel)
	slog.Info("starting application", "environment", env)

	server.Init(conf)
}
