package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Buckets are tuned to the Battlesnake move deadline (500ms), and dense between
// 200ms and 500ms because that is where a searching bot actually lives.
var requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "battlesnake_request_duration_seconds",
	Help: "HTTP request latency in seconds, by route.",
	Buckets: []float64{
		.005, .01, .025, .05, .1, .15, // the cheap endpoints: info, start, end
		.2, .22, .24, .26, .28, .3, .32, .34, .36, .38, .4, // where moves live
		.45, .5, .75, 1,
	},
}, []string{"route", "bot", "method", "status"})

var gamesFinished = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "battlesnake_games_finished_total",
	Help: "Games finished, by bot and result.",
}, []string{"bot", "result"})

func Handler() http.Handler { return promhttp.Handler() }

// InitGameCounters registers every bot/result series at zero. Without it a
// series first appears with value 1, increase() never sees the 0->1 rise, and
// win rate reads NaN until the second game. It also makes the panel show 0%
// rather than "No data" before any game finishes.
func InitGameCounters(bots map[string]bool) {
	for bot := range bots {
		gamesFinished.WithLabelValues(bot, "win").Add(0)
		gamesFinished.WithLabelValues(bot, "loss").Add(0)
	}
}

func ObserveGameFinished(bot string, win bool) {
	result := "loss"
	if win {
		result = "win"
	}
	gamesFinished.WithLabelValues(bot, result).Inc()
}

func Observe(route, bot, method string, status int, took time.Duration) {
	requestDuration.WithLabelValues(route, bot, method, strconv.Itoa(status)).Observe(took.Seconds())
}

func Classify(path string, knownBots map[string]bool) (route, bot string) {
	trimmed := strings.Trim(path, "/")
	switch trimmed {
	case "":
		return "/", ""
	case "metrics":
		return "/metrics", ""
	}

	parts := strings.SplitN(trimmed, "/", 3)
	if !knownBots[parts[0]] {
		return "other", "unknown"
	}
	bot = parts[0]

	if len(parts) == 1 {
		return "/{bot}", bot
	}
	if len(parts) == 2 {
		switch parts[1] {
		case "start", "end", "move":
			return "/{bot}/" + parts[1], bot
		}
	}
	return "other", bot
}
