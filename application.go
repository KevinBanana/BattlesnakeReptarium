package main

import (
	"log/slog"
	"os"

	"BattlesnakeReptarium/internal/config"
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

	var lvl slog.Level
	_ = lvl.UnmarshalText([]byte(conf.LogLevel))
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})))

	slog.Info("starting application", "environment", env)

	server.Init(conf)
}
