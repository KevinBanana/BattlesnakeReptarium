package controllers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"

	slogctx "github.com/veqryn/slog-context"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/services"
)

type GameController struct {
	bots          map[string]services.Bot
	gameEngineSvc services.GameEngineService
}

func NewGameController(bots map[string]services.Bot, gameEngineSvc services.GameEngineService) GameController {
	return GameController{
		bots:          bots,
		gameEngineSvc: gameEngineSvc,
	}
}

type botHandler func(w http.ResponseWriter, r *http.Request, bot services.Bot)

// WithBot resolves the {bot} path segment before handing off, 404ing if the
// name is unknown.
func (g GameController) WithBot(h botHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("bot")

		bot, ok := g.bots[name]
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Errorf("unknown bot %q", name))
			return
		}

		ctx := slogctx.Prepend(r.Context(), "bot", name)
		h(w, r.WithContext(ctx), bot)
	}
}

func (g GameController) StartGame(w http.ResponseWriter, r *http.Request, _ services.Bot) {
	var reqBody model.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	ctx := slogctx.Prepend(r.Context(), "game", reqBody.Game.ID)

	if err := g.gameEngineSvc.StartGame(ctx, reqBody.Game, reqBody.Board, reqBody.SelfSnake); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	slog.InfoContext(ctx, "game started", "ruleset", reqBody.Game.Ruleset.Name, "map", reqBody.Game.Map)

	w.WriteHeader(http.StatusOK)
}

func (g GameController) EndGame(w http.ResponseWriter, r *http.Request, _ services.Bot) {
	var reqBody model.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	ctx := slogctx.Prepend(r.Context(), "game", reqBody.Game.ID)

	if err := g.gameEngineSvc.EndGame(ctx, reqBody.Game, reqBody.Board, reqBody.SelfSnake); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	slog.InfoContext(ctx, "game ended",
		"win", reqBody.Game.IsWin,
		"turns", reqBody.Turn,
		"length", reqBody.SelfSnake.Length,
		"health", reqBody.SelfSnake.Health,
	)

	w.WriteHeader(http.StatusOK)
}

func (g GameController) CalculateMove(w http.ResponseWriter, r *http.Request, bot services.Bot) {
	var reqBody model.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	ctx := slogctx.Prepend(r.Context(), "game", reqBody.Game.ID, "turn", reqBody.Turn)

	slog.InfoContext(ctx, "move request received",
		"health", reqBody.SelfSnake.Health,
		"length", reqBody.SelfSnake.Length,
		"head", reqBody.SelfSnake.Head,
		"snakes", len(reqBody.Board.Snakes),
	)
	slog.DebugContext(ctx, "board state", "board", reqBody.Board)

	snakeAction, err := bot.CalculateMove(ctx, reqBody.Game, reqBody.Turn, reqBody.Board, reqBody.SelfSnake)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	slog.InfoContext(ctx, "move response", "action", snakeAction)

	writeJSON(w, http.StatusOK, map[string]any{
		"move":  snakeAction.Move,
		"shout": snakeAction.Shout,
	})
}

// Info returns metadata for a specific snake bot. apiversion and author are
// the same for every snake, only the customizations differ.
func (g GameController) Info(w http.ResponseWriter, _ *http.Request, bot services.Bot) {
	customizations := bot.Customizations()

	writeJSON(w, http.StatusOK, map[string]string{
		"apiversion": "1",
		"author":     "Kevin Bonanno",
		"color":      customizations.Color,
		"head":       customizations.Head,
		"tail":       customizations.Tail,
	})
}

func (g GameController) Health(w http.ResponseWriter, r *http.Request) {
	bots := slices.Sorted(maps.Keys(g.bots))

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"bots":   bots,
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
