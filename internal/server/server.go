package server

import (
	"fmt"

	"BattlesnakeReptarium/internal/config"
	"BattlesnakeReptarium/internal/controllers"
	"BattlesnakeReptarium/internal/repo"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
	controller := controllers.NewGameController(&db, conf.ActiveBot)
	r := NewRouter(controller)

	listenAddress := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	log.Info(fmt.Sprintf("Listening on %s", listenAddress))
	log.Fatal(r.Run(listenAddress))
}
