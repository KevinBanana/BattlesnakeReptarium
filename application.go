package main

import (
	"log/slog"
	"os"

	slogctx "github.com/veqryn/slog-context"

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

	handler := slogctx.NewHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}), nil)
	slog.SetDefault(slog.New(handler))

	slog.Info("starting application", "env", env)

	server.Init(conf)
}
