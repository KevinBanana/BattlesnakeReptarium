// Command selfplay plays a batch of constrictor games and writes what the loop
// needs: training records, a rendered game, and a summary.
//
// Self-play, every seat the same network:
//
//	go run ./cmd/selfplay -model runs/dev/champion.onnx -games 2000 \
//	    -samples runs/dev/iter-0007/games.bin -html runs/dev/iter-0007/game.html
//
// Evaluation, one seat against something else, rotated through every start
// position. -against takes a model path or the word "voronoi":
//
//	go run ./cmd/selfplay -model candidate.onnx -against champion.onnx \
//	    -games 400 -json eval.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"BattlesnakeReptarium/internal/sim/constrictor"
	"BattlesnakeReptarium/internal/sim/nn"
	"BattlesnakeReptarium/internal/sim/selfplay"
)

func main() {
	var (
		model    = flag.String("model", "", "ONNX model to play (required)")
		against  = flag.String("against", "", `opponent: a model path, or "voronoi". Empty means self-play`)
		board    = flag.Int("board", 11, "board size")
		seats    = flag.Int("seats", 4, "snakes per game")
		games    = flag.Int("games", 100, "games to play")
		sims     = flag.Int("sims", 200, "search simulations per move")
		workers  = flag.Int("workers", runtime.NumCPU(), "games in parallel")
		seed     = flag.Int64("seed", time.Now().UnixNano(), "rng seed")
		samples  = flag.String("samples", "", "write training records here (self-play only)")
		html     = flag.String("html", "", "render the first game here")
		summary  = flag.String("json", "", "write a summary here")
		cuda     = flag.Bool("cuda", false, "run inference on the GPU (needs the GPU build of the ONNX Runtime library)")
		batch    = flag.Int("batch", 0, "batch evaluations from concurrent games into calls of up to this many positions; 0 calls the session directly")
		linger   = flag.Duration("linger", 500*time.Microsecond, "how long a batch keeps gathering before it goes")
		temp     = flag.Float64("temperature", 1.0, "self-play move sampling temperature")
		until    = flag.Int("temperature-until", 8, "turn after which the best move is always taken")
		noise    = flag.Float64("noise", 0.25, "Dirichlet noise mixed into the root prior in self-play")
		progress = flag.Int("progress", 25, "log every N games; 0 to stay quiet")
	)
	flag.Parse()

	if *model == "" {
		log.Fatal("-model is required")
	}
	if *samples != "" && *against != "" {
		log.Fatal("-samples is for self-play: records from a game against a different opponent are not the network's own moves")
	}

	log.SetFlags(log.Ltime)
	if err := run(runConfig{
		model: *model, against: *against,
		board: *board, seats: *seats, games: *games, sims: *sims,
		workers: *workers, seed: *seed,
		samples: *samples, html: *html, summary: *summary,
		temperature: *temp, until: *until, noise: *noise, progress: *progress,
		cuda: *cuda, batch: *batch, linger: *linger,
	}); err != nil {
		log.Fatal(err)
	}
}

type runConfig struct {
	model, against            string
	board, seats, games, sims int
	workers                   int
	seed                      int64
	samples, html, summary    string
	temperature               float64
	until                     int
	noise                     float64
	progress                  int
	cuda                      bool
	batch                     int
	linger                    time.Duration
}

// openModel loads a model and returns what each worker's evaluator should call:
// the session directly, or its own client on a batching server.
//
// The server is only worth its latency where a big call costs about what a
// small one does, which in practice means the GPU. Batches come from games in
// flight, so -batch wants -workers of at least a quarter of it - each worker
// contributes up to four rows, one per living seat.
func openModel(cfg runConfig, path string) (newRunner func() nn.Runner, closer func(), err error) {
	session, err := nn.OpenWith(path, constrictor.Planes, cfg.board, cfg.board, nn.Options{CUDA: cfg.cuda})
	if err != nil {
		return nil, nil, err
	}
	if cfg.batch <= 0 {
		return func() nn.Runner { return session }, func() { session.Close() }, nil
	}

	server := nn.NewServer(session, cfg.batch, cfg.workers*2, cfg.linger)
	return func() nn.Runner { return server.Client() },
		func() {
			server.Close()
			if calls, rows := server.Stats(); calls > 0 {
				log.Printf("%s: %d positions in %d calls, average batch %.1f",
					filepath.Base(path), rows, calls, float64(rows)/float64(calls))
			}
			session.Close()
		}, nil
}

func run(cfg runConfig) error {
	newRunner, closeModel, err := openModel(cfg, cfg.model)
	if err != nil {
		return err
	}
	defer closeModel()

	// Evaluation plays its best move; self-play explores.
	playerCfg := nn.PlayerConfig{Sims: cfg.sims}
	if cfg.against == "" {
		playerCfg.Temperature = cfg.temperature
		playerCfg.UntilTurn = cfg.until
		playerCfg.DirichletNoise = cfg.noise
	}

	// Each worker gets its own rng, evaluator and runner; the model is shared.
	subject := func(worker int) selfplay.Searcher {
		return nn.NewPlayer(newRunner(), cfg.board, playerCfg, rand.New(rand.NewSource(cfg.seed+int64(worker)*104729)))
	}

	batch := selfplay.Batch{
		Board: cfg.board, Seats: cfg.seats, Games: cfg.games,
		Workers: cfg.workers, Seed: cfg.seed,
		Progress: progressLogger(cfg.games, cfg.progress),
	}

	started := time.Now()
	var result selfplay.Result

	switch cfg.against {
	case "":
		log.Printf("self-play: %d games, %d sims/move, %d workers, board %d", cfg.games, cfg.sims, cfg.workers, cfg.board)
		result = selfplay.SelfPlay(batch, subject)

	case "voronoi":
		log.Printf("evaluation vs voronoi: %d games, %d sims/move", cfg.games, cfg.sims)
		result = selfplay.Evaluate(batch, subject, func(int) selfplay.Searcher { return selfplay.VoronoiSearcher{} })

	default:
		newOpponent, closeOpponent, err := openModel(cfg, cfg.against)
		if err != nil {
			return err
		}
		defer closeOpponent()

		log.Printf("evaluation vs %s: %d games, %d sims/move", filepath.Base(cfg.against), cfg.games, cfg.sims)
		result = selfplay.Evaluate(batch, subject, func(worker int) selfplay.Searcher {
			return nn.NewPlayer(newOpponent(), cfg.board, nn.PlayerConfig{Sims: cfg.sims},
				rand.New(rand.NewSource(cfg.seed+int64(worker)*15485863)))
		})
	}

	elapsed := time.Since(started)
	log.Printf("%d games in %s (%.1f games/s), %d samples, average placement %+.3f, finishes %v",
		result.Games, elapsed.Round(time.Millisecond), float64(result.Games)/elapsed.Seconds(),
		len(result.Samples), result.Average(), result.Ranks[:cfg.seats])

	return writeOutputs(cfg, result, elapsed)
}

func writeOutputs(cfg runConfig, result selfplay.Result, elapsed time.Duration) error {
	if cfg.samples != "" {
		if err := writeSamples(cfg, result); err != nil {
			return err
		}
	}

	if cfg.html != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.html), 0o750); err != nil {
			return err
		}
		// Ego 0 in self-play is an arbitrary pick - every seat is the same
		// weights - but one blue snake to follow is the difference between a
		// watchable game and four identical worms.
		if err := os.WriteFile(cfg.html, []byte(constrictor.HTML(result.Frames, 0)), 0o600); err != nil {
			return err
		}
		log.Printf("wrote %s (%d turns)", cfg.html, len(result.Frames)-1)
	}

	if cfg.summary != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.summary), 0o750); err != nil {
			return err
		}
		body, err := json.MarshalIndent(map[string]any{
			"model":      cfg.model,
			"against":    cfg.against,
			"games":      result.Games,
			"sims":       cfg.sims,
			"samples":    len(result.Samples),
			"placement":  result.Average(),
			"finishes":   result.Ranks[:cfg.seats],
			"seconds":    elapsed.Seconds(),
			"turns_seen": len(result.Frames) - 1,
		}, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(cfg.summary, body, 0o600); err != nil {
			return err
		}
		log.Printf("wrote %s", cfg.summary)
	}
	return nil
}

func writeSamples(cfg runConfig, result selfplay.Result) error {
	if err := os.MkdirAll(filepath.Dir(cfg.samples), 0o750); err != nil {
		return err
	}

	// Written to a temporary name and renamed, so an interrupted run never
	// leaves a half-file that looks like a finished iteration.
	tmp := cfg.samples + ".partial"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	w, err := constrictor.NewSampleWriter(f, cfg.board, cfg.board, cfg.seats)
	if err != nil {
		f.Close()
		return err
	}
	if err := w.Write(result.Samples); err != nil {
		f.Close()
		return err
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, cfg.samples); err != nil {
		return err
	}

	log.Printf("wrote %s (%d samples)", cfg.samples, len(result.Samples))
	return nil
}

func progressLogger(total, every int) func(int) {
	if every <= 0 {
		return nil
	}
	started := time.Now()
	return func(done int) {
		if done%every != 0 && done != total {
			return
		}
		elapsed := time.Since(started)
		rate := float64(done) / elapsed.Seconds()
		eta := time.Duration(float64(total-done)/rate) * time.Second
		fmt.Fprintf(os.Stderr, "\r  %d/%d games  %.1f/s  eta %s        ", done, total, rate, eta.Round(time.Second))
		if done == total {
			fmt.Fprintln(os.Stderr)
		}
	}
}
