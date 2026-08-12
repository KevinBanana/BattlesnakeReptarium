package server

import (
	"fmt"
	"log/slog"
	"os"

	"BattlesnakeReptarium/internal/config"
	"BattlesnakeReptarium/internal/controllers"
	"BattlesnakeReptarium/internal/repo"
	"BattlesnakeReptarium/internal/services"
	"BattlesnakeReptarium/internal/services/bananabot"
	"BattlesnakeReptarium/internal/services/bananatron"

	"github.com/gin-gonic/gin"
)

const (
	botBananaBotV1 = "banana_bot_v1"
	botBananatron  = "bananatron"
)

func NewRouter(controller controllers.GameController) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.POST("/start", controller.StartGame)
	router.POST("/end", controller.EndGame)
	router.POST("/move", controller.CalculateMove)
	router.GET("/", controller.Health)

	slog.Info("router created")
	return router
}

func Init(conf *config.Config) {
	db := repo.NewDatabase()
	botSvc := newBotService(conf.ActiveBot)
	gameEngineSvc := services.NewGameEngineSvc(db)
	controller := controllers.NewGameController(*botSvc, gameEngineSvc)
	r := NewRouter(controller)

	listenAddress := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	slog.Info("listening", "address", listenAddress)
	if err := r.Run(listenAddress); err != nil {
		slog.Error("server exited", "err", err)
		os.Exit(1)
	}
}

func newBotService(activeBot string) *services.Bot {
	var botSvc services.Bot

	switch activeBot {
	case botBananaBotV1:
		botSvc = bananabot.New()
	case botBananatron:
		botSvc = bananatron.New()
	default:
		slog.Error("cannot set active bot: not found", "activeBot", activeBot)
		return nil
	}

	return &botSvc
}
