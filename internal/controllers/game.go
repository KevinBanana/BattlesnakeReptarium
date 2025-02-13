package controllers

import (
	"net/http"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type GameController struct {
	bot           services.Bot
	gameEngineSvc services.GameEngineService
}

func NewGameController(botSvc services.Bot, gameEngineSvc services.GameEngineService) GameController {
	return GameController{
		bot:           botSvc,
		gameEngineSvc: gameEngineSvc,
	}
}

func (g GameController) StartGame(ctx *gin.Context) {
	var reqBody model.RequestBody
	if err := ctx.ShouldBindBodyWith(&reqBody, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	err := g.gameEngineSvc.StartGame(ctx, reqBody.Game, reqBody.Board, reqBody.SelfSnake)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.Error{Err: err})
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func (g GameController) EndGame(ctx *gin.Context) {
	var reqBody model.RequestBody
	if err := ctx.ShouldBindBodyWith(&reqBody, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	err := g.gameEngineSvc.EndGame(ctx, reqBody.Game, reqBody.Board, reqBody.SelfSnake)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.Error{Err: err})
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func (g GameController) CalculateMove(ctx *gin.Context) {
	if g.bot == nil {
		ctx.JSON(http.StatusInternalServerError, gin.Error{
			Err: errors.New("bot not set"),
		})
		return
	}

	var reqBody model.RequestBody
	if err := ctx.ShouldBindBodyWith(&reqBody, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	log.WithField("reqBody", reqBody).Info("move request received")

	snakeAction, err := g.bot.CalculateMove(ctx, reqBody.Game, reqBody.Turn, reqBody.Board, reqBody.SelfSnake)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.Error{Err: err})
		return
	}

	log.WithField("move response", snakeAction)

	ctx.JSON(http.StatusOK, gin.H{
		"move":  snakeAction.Move,
		"shout": snakeAction.Shout,
	})
}

func (g GameController) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"apiversion": "1",
		"author":     "Kevin Bonanno",
		"color":      "#e8f008",
		"head":       "fang",
		"tail":       "round-bum",
	})
}
