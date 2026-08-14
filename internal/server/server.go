package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	slogctx "github.com/veqryn/slog-context"

	"BattlesnakeReptarium/internal/config"
	"BattlesnakeReptarium/internal/controllers"
	"BattlesnakeReptarium/internal/repo"
	"BattlesnakeReptarium/internal/services"
	"BattlesnakeReptarium/internal/services/bananabot"
	"BattlesnakeReptarium/internal/services/bananatron"
)

func NewRouter(controller controllers.GameController) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", controller.Health)
	mux.HandleFunc("GET /{bot}", controller.WithBot(controller.Info))
	mux.HandleFunc("GET /{bot}/{$}", controller.WithBot(controller.Info))

	mux.HandleFunc("POST /{bot}/start", controller.WithBot(controller.StartGame))
	mux.HandleFunc("POST /{bot}/end", controller.WithBot(controller.EndGame))
	mux.HandleFunc("POST /{bot}/move", controller.WithBot(controller.CalculateMove))

	slog.Info("router created")
	return logRequests(mux)
}

func Init(conf *config.Config) {
	db := repo.NewDatabase()
	gameEngineSvc := services.NewGameEngineSvc(db)
	controller := controllers.NewGameController(newBots(), gameEngineSvc)
	handler := NewRouter(controller)

	listenAddress := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	slog.Info("listening", "address", listenAddress)

	srv := &http.Server{
		Addr:              listenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server exited", "err", err)
		os.Exit(1)
	}
}

// newBots returns every bot, keyed by the path segment it is served under.
func newBots() map[string]services.Bot {
	return map[string]services.Bot{
		"bananabot":  bananabot.New(),
		"bananatron": bananatron.New(),
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := slogctx.Prepend(r.Context(), "method", r.Method, "path", r.URL.Path)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r.WithContext(ctx))

		slog.InfoContext(ctx, "request",
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
