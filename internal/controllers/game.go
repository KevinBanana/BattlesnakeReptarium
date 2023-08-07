package controllers

import (
	"net/http"
	"time"

	"BattlesnakeReptarium/internal/repo"
	"BattlesnakeReptarium/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type GameController struct {
	bot           services.Bot
	gameEngineSvc services.GameEngineSvc
}

func NewGameController(db repo.DB, activeBot string) GameController {
	botSvc := createSelectedBot(activeBot)
	gameEngineSvc := services.NewGameEngineSvc(db)

	return GameController{
		bot:           *botSvc,
		gameEngineSvc: *gameEngineSvc,
	}
}

func createSelectedBot(activeBot string) *services.Bot {
	var botSvc services.Bot

	switch activeBot {
	case "banana_bot_v1":
		botSvc = services.NewBananaBotV1Svc()
	}

	return &botSvc
}

func (g GameController) StartGame(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}

func (g GameController) EndGame(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}

func (g GameController) CalculateMove(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}

func (g GameController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"time": time.Now()})
}
