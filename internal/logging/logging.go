package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Init installs a JSON slog logger as the process-wide default at the given
// level. slog has no Trace or Fatal levels, so trace folds into Debug and fatal
// into Error; anything unrecognized falls back to Info.
func Init(level string) {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "trace", "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error", "fatal":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	slog.SetDefault(slog.New(handler))
}
