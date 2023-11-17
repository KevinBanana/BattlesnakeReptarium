package main

import (
	"fmt"
	"os"

	"BattlesnakeReptarium/internal/config"
	"BattlesnakeReptarium/internal/logging"
	"BattlesnakeReptarium/internal/server"

	log "github.com/sirupsen/logrus"
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
	log.Info(fmt.Sprintf("Starting application in %s environment", env))
	config.Init(env)
	logging.SetLogLevel(config.GetConfig().LogLevel)
	log.SetFormatter(&log.JSONFormatter{})

	server.Init()
}
