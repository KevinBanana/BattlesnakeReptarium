package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"BattlesnakeReptarium/internal/config"
	"BattlesnakeReptarium/internal/controllers"
	"BattlesnakeReptarium/internal/repo"
	"BattlesnakeReptarium/internal/services"
	"BattlesnakeReptarium/internal/services/bananabot"
	"BattlesnakeReptarium/internal/services/bananatron"
)

const (
	botBananaBotV1 = "banana_bot_v1"
	botBananatron  = "bananatron"
)

func NewRouter(controller controllers.GameController) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /start", controller.StartGame)
	mux.HandleFunc("POST /end", controller.EndGame)
	mux.HandleFunc("POST /move", controller.CalculateMove)
	mux.HandleFunc("GET /{$}", controller.Health)

	slog.Info("router created")
	return logRequests(mux)
}

func Init(conf *config.Config) {
	db := repo.NewDatabase()
	botSvc := newBotService(conf.ActiveBot)
	gameEngineSvc := services.NewGameEngineSvc(db)
	controller := controllers.NewGameController(*botSvc, gameEngineSvc)
	handler := NewRouter(controller)

	listenAddress := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	slog.Info("listening", "address", listenAddress)
	if err := http.ListenAndServe(listenAddress, handler); err != nil {
		slog.Error("server exited", "err", err)
		os.Exit(1)
	}
}

func newBotService(activeBot string) *services.Bot {
	var botSvc services.Bot

	switch activeBot {
	case botBananaBotV1:
		botSvc = bananabot.New()
	case botBananatron:
		botSvc = bananatron.New()
	default:
		slog.Error("cannot set active bot: not found", "activeBot", activeBot)
		return nil
	}

	return &botSvc
}

// logRequests logs one line per request
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
