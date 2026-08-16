package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strconv"

	slogctx "github.com/veqryn/slog-context"

	"BattlesnakeReptarium/internal/metrics"
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
type gameHandler func(ctx context.Context, w http.ResponseWriter, body model.RequestBody, name string, bot services.Bot)

// WithBot resolves the {bot} path segment before handing off, 404ing if the
// name is unknown.
func (g GameController) WithBot(h botHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("bot")

		bot, ok := g.bots[name]
		if !ok {
			slog.WarnContext(r.Context(), "unknown bot", "bot", name)
			writeStatus(w, http.StatusNotFound)
			return
		}

		h(w, r, bot)
	}
}

func (g GameController) WithGame(h gameHandler) http.HandlerFunc {
	return g.WithBot(func(w http.ResponseWriter, r *http.Request, bot services.Bot) {
		name := r.PathValue("bot")

		var body model.RequestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			slog.WarnContext(r.Context(), "invalid request body", "bot", name, "err", err)
			writeStatus(w, http.StatusBadRequest)
			return
		}

		ctx := slogctx.Prepend(r.Context(), "bot", name, "game", body.Game.ID, "turn", body.Turn)
		h(ctx, w, body, name, bot)
	})
}

func (g GameController) StartGame(ctx context.Context, w http.ResponseWriter, body model.RequestBody, _ string, bot services.Bot) {
	if !services.PlaysGamemode(bot, body.Game.Ruleset.Name) {
		slog.WarnContext(ctx, "bot entered a gamemode it is not built for",
			"ruleset", body.Game.Ruleset.Name, "supports", bot.Gamemodes())
	}

	if err := g.gameEngineSvc.StartGame(ctx, body.Game, body.Board, body.SelfSnake); err != nil {
		slog.ErrorContext(ctx, "failed to start game", "err", err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "game started", "ruleset", body.Game.Ruleset.Name, "map", body.Game.Map)

	w.WriteHeader(http.StatusOK)
}

func (g GameController) EndGame(ctx context.Context, w http.ResponseWriter, body model.RequestBody, name string, _ services.Bot) {
	finished, err := g.gameEngineSvc.EndGame(ctx, body.Game, body.Board, body.SelfSnake)
	if err != nil {
		slog.ErrorContext(ctx, "failed to end game", "err", err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}

	metrics.ObserveGameFinished(name, finished.IsWin)

	slog.InfoContext(ctx, "game ended",
		"win", finished.IsWin,
		"length", body.SelfSnake.Length,
		"health", body.SelfSnake.Health,
	)

	w.WriteHeader(http.StatusOK)
}

func (g GameController) CalculateMove(ctx context.Context, w http.ResponseWriter, body model.RequestBody, _ string, bot services.Bot) {
	slog.InfoContext(ctx, "move request received",
		"health", body.SelfSnake.Health,
		"length", body.SelfSnake.Length,
		"head", body.SelfSnake.Head,
		"snakes", len(body.Board.Snakes),
		// What the battlesnake server measured for our previous move, in milliseconds. The
		// gap between this and the duration we log for that request is
		// everything outside this process including the round trip
		"engine_latency_ms", latencyMillis(body.SelfSnake.Latency),
		"timeout_ms", body.Game.Timeout,
	)
	slog.DebugContext(ctx, "board state", "board", body.Board)

	snakeAction, err := bot.CalculateMove(ctx, body.Game, body.Turn, body.Board, body.SelfSnake)
	if err != nil {
		slog.ErrorContext(ctx, "failed to calculate move", "err", err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "move response", "action", snakeAction)

	writeJSON(w, http.StatusOK, map[string]any{
		"move":  snakeAction.Move,
		"shout": snakeAction.Shout,
	})
}

// latencyMillis parses the engine's latency field, which arrives as a string of
// milliseconds and is empty on the first move of a game. Logged as a number so
// it can be compared and aggregated rather than only matched.
func latencyMillis(latency string) int {
	ms, err := strconv.Atoi(latency)
	if err != nil {
		return 0
	}
	return ms
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

// writeStatus replies with the status text only. The snake is public via
// Funnel, so internal error detail stays in the log and never reaches the
// client.
func writeStatus(w http.ResponseWriter, status int) {
	writeJSON(w, status, map[string]string{"error": http.StatusText(status)})
}
