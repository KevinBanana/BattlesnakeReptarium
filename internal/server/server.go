package server

import (
	"fmt"

	"BattlesnakeReptarium/internal/config"
	"BattlesnakeReptarium/internal/controllers"
	"BattlesnakeReptarium/internal/repo"
	"BattlesnakeReptarium/internal/services"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	bananaBotV1 = "banana_bot_v1"
)

func NewRouter(controller controllers.GameController) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.POST("/start", controller.StartGame)
	router.POST("/end", controller.EndGame)
	router.POST("/move", controller.CalculateMove)
	router.GET("/", controller.Health)

	log.Info("Router created")
	return router
}

func Init() {
	conf := config.GetConfig()
	db := repo.Database{}
	botSvc := createSelectedBotService(conf.ActiveBot)
	controller := controllers.NewGameController(&db, *botSvc)
	r := NewRouter(controller)

	listenAddress := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	log.Info(fmt.Sprintf("Listening on %s", listenAddress))
	log.Fatal(r.Run(listenAddress))
}

func createSelectedBotService(activeBot string) *services.Bot {
	var botSvc services.Bot

	switch activeBot {
	case bananaBotV1:
		botSvc = services.NewBananaBotV1Svc()
	default:
		log.Errorf("Cannot set active bot: '%s' not found", activeBot)
		return nil
	}

	return &botSvc
}
