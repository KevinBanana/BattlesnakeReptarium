package server

import (
	"fmt"

	"BattlesnakeReptarium/internal/config"
	"BattlesnakeReptarium/internal/controllers"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	gameEngineGroup := router.Group("")
	{
		gameEngine := new(controllers.GameController)
		gameEngineGroup.POST("/start", gameEngine.StartGame)
		gameEngineGroup.POST("/end", gameEngine.EndGame)
		gameEngineGroup.POST("/move", gameEngine.CalculateMove)
		gameEngineGroup.GET("/", gameEngine.Health)
	}

	log.Info("Router created")
	return router
}

func Init() {
	r := NewRouter()

	conf := config.GetConfig()
	listenAddress := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	log.Info(fmt.Sprintf("Listening on %s", listenAddress))
	log.Fatal(r.Run(listenAddress))
}
