package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/services"
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

func (g GameController) StartGame(w http.ResponseWriter, r *http.Request) {
	var reqBody model.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := g.gameEngineSvc.StartGame(r.Context(), reqBody.Game, reqBody.Board, reqBody.SelfSnake); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g GameController) EndGame(w http.ResponseWriter, r *http.Request) {
	var reqBody model.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := g.gameEngineSvc.EndGame(r.Context(), reqBody.Game, reqBody.Board, reqBody.SelfSnake); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g GameController) CalculateMove(w http.ResponseWriter, r *http.Request) {
	if g.bot == nil {
		writeError(w, http.StatusInternalServerError, errors.New("bot not set"))
		return
	}

	var reqBody model.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	slog.Info("move request received", "reqBody", reqBody)

	snakeAction, err := g.bot.CalculateMove(r.Context(), reqBody.Game, reqBody.Turn, reqBody.Board, reqBody.SelfSnake)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	slog.Info("move response", "action", snakeAction)

	writeJSON(w, http.StatusOK, map[string]any{
		"move":  snakeAction.Move,
		"shout": snakeAction.Shout,
	})
}

func (g GameController) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"apiversion": "1",
		"author":     "Kevin Bonanno",
		"color":      "#e8f008",
		"head":       "fang",
		"tail":       "round-bum",
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
