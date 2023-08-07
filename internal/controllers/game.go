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
	gameEngine services.GameEngineSvc
	//bananaBotV1 services.
}

func NewGameController(db repo.DB) GameController {
	gameEngineSvc := services.NewGameEngineSvc(db)
	return GameController{
		gameEngine: *gameEngineSvc,
	}
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
	// TODO set active bot from config
	c.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}

func (g GameController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"time": time.Now()})
}
