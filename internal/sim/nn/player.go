package nn

import (
	"math/rand"

	"BattlesnakeReptarium/internal/sim/constrictor"
)

// Player is the network with search in front of it: what self-play runs, what
// evaluation measures, and what Phase 4 will serve.
//
// It deliberately does not import the selfplay package. Search is the only
// method anything needs, and Go matches that structurally, so the dependency
// runs one way - selfplay knows nothing about ONNX Runtime, and stays buildable
// without a C toolchain.
type Player struct {
	search      constrictor.Search
	temperature float64
	untilTurn   int
	rnd         *rand.Rand
}

// PlayerConfig tunes how a network plays.
type PlayerConfig struct {
	Sims int

	// Temperature and UntilTurn shape self-play exploration: moves are drawn
	// from the visit distribution while Turn < UntilTurn, and the most-visited
	// move is taken after. Games here are short and the opening is a large
	// fraction of one, so exploration has to be front-loaded and then cut.
	// Evaluation leaves both zero and always plays the best move.
	Temperature float64
	UntilTurn   int

	// DirichletNoise mixes exploration noise into the root prior. Self-play
	// wants it; evaluation wants the policy as it is.
	DirichletNoise float64
}

func NewPlayer(runner Runner, board int, cfg PlayerConfig, rnd *rand.Rand) *Player {
	return &Player{
		search: constrictor.Search{
			Sims:           cfg.Sims,
			Eval:           NewEvaluator(runner, board, board),
			Rnd:            rnd,
			DirichletNoise: cfg.DirichletNoise,
		},
		temperature: cfg.Temperature,
		untilTurn:   cfg.UntilTurn,
		rnd:         rnd,
	}
}

// Search returns every seat's move and the visit distribution behind it. The
// distribution is the policy target regardless of which move was played - what
// is worth learning is what search concluded, not what exploration sampled.
func (p *Player) Search(s *constrictor.State) (moves [constrictor.MaxSnakes]constrictor.Move, policy [constrictor.MaxSnakes][4]float32) {
	result := p.search.Run(s)

	temperature := 0.0
	if s.Turn < p.untilTurn {
		temperature = p.temperature
	}

	for seat := 0; seat < s.N; seat++ {
		if !s.IsAlive(seat) {
			continue
		}
		policy[seat] = result.Policy(seat)
		moves[seat] = result.Sample(seat, temperature, p.rnd)
	}
	return moves, policy
}
