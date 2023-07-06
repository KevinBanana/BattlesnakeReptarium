package application

import (
	"context"
	"os"
	"time"

	"BattlesnakeReptarium/internal"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type App struct {
	srv *gin.Engine
}

func New() *App {
	application := App{}

	conf := internal.Config{Environment: getEnvironment()}
	err := conf.GetConfig()
	if err != nil {
		log.Errorf("New::conf.GetConfig: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return &application
}

func (a *App) Run() {
	err := a.srv.Run()
	if err != nil {
		log.Fatalf("Run::srv.Run: %v", err)
	}
}

func getEnvironment() string {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		return "local"
	}
	return environment
}
