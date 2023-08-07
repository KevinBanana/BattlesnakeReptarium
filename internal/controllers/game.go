package controllers

import (
	"net/http"
	"time"

	"BattlesnakeReptarium/internal/repo"
	"BattlesnakeReptarium/internal/services"
	"BattlesnakeReptarium/internal/services/bots"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type GameController struct {
	bananaBotV1   bots.BananaBotV1Svc
	gameEngineSvc services.GameEngineSvc
}

func NewGameController(db repo.DB) GameController {
	bananaBotV1Svc := bots.NewBananaBotV1Svc()
	gameEngineSvc := services.NewGameEngineSvc(db)

	return GameController{
		bananaBotV1:   *bananaBotV1Svc,
		gameEngineSvc: *gameEngineSvc,
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
	c.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}

func (g GameController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"time": time.Now()})
}
