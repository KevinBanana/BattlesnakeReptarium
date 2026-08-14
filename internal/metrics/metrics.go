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

// Buckets are tuned to the Battlesnake move deadline (500ms). The client
// default tops out at 10s, which would put every real move in one bucket.
var requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "battlesnake_request_duration_seconds",
	Help:    "HTTP request latency in seconds, by route.",
	Buckets: []float64{.001, .0025, .005, .01, .025, .05, .1, .25, .5, 1},
}, []string{"route", "bot", "method", "status"})

func Handler() http.Handler { return promhttp.Handler() }

func Observe(route, bot, method string, status int, took time.Duration) {
	requestDuration.WithLabelValues(route, bot, method, strconv.Itoa(status)).Observe(took.Seconds())
}

func Classify(path string, knownBots map[string]bool) (route, bot string) {
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
