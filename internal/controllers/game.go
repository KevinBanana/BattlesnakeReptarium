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

	var err error
	// Get the game, turn, board, and self from the request
	var game model.Game
	if err = bindBody(ctx, &game); err != nil {
		return
	}
	var turn int
	if err = bindBody(ctx, &turn); err != nil {
		return
	}
	var board model.Board
	if err = bindBody(ctx, &board); err != nil {
		return
	}
	var selfSnake model.Snake
	if err = bindBody(ctx, &selfSnake); err != nil {
		return
	}

	snakeMove, err := g.bot.CalculateMove(ctx, game, turn, board, selfSnake)
	ctx.JSON(http.StatusInternalServerError, gin.Error{
		Err: errors.New("not implemented"),
	})
}

func (g GameController) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"time": time.Now()})
}

func bindBody(ctx *gin.Context, obj interface{}) error {
	if err := ctx.ShouldBindBodyWith(obj, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	return nil
}
