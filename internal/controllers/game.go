package controllers

import (
	"net/http"
	"time"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/repo"
	"BattlesnakeReptarium/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type GameController struct {
	bot           services.Bot
	gameEngineSvc services.GameEngineSvc
}

const (
	bananaBotV1 = "banana_bot_v1"
)

func NewGameController(db repo.DB, activeBot string) GameController {
	botSvc := createSelectedBotService(activeBot)
	gameEngineSvc := services.NewGameEngineSvc(db)

	return GameController{
		bot:           *botSvc,
		gameEngineSvc: *gameEngineSvc,
	}
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

func (g GameController) StartGame(ctx *gin.Context) {
	ctx.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}

func (g GameController) EndGame(ctx *gin.Context) {
	ctx.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}

func (g GameController) CalculateMove(ctx *gin.Context) {
	if g.bot == nil {
		ctx.JSON(http.StatusInternalServerError, gin.Error{
			Err: errors.New("bot not set"),
		})
		return
	}

	var reqBody model.MoveRequestBody
	// Get the game, turn, board, and self from the request
	if err := ctx.ShouldBindBodyWith(&reqBody, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	snakeAction, err := g.bot.CalculateMove(ctx, game, turn, board, selfSnake)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.Error{Err: err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"move":  snakeAction.Move,
		"shout": snakeAction.Shout,
	})
}

func (g GameController) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"time": time.Now()})
}
