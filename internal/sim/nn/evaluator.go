package nn

import (
	"math"

	"BattlesnakeReptarium/internal/sim/constrictor"
)

// Runner is where an evaluation actually happens: straight to a Session, or
// through a Server that batches it with other games' positions. Evaluator
// cannot tell the two apart.
type Runner interface {
	Run(in []float32, batch int) (policy, value []float32, err error)
}

// Evaluator plugs a network into search as constrictor.Evaluator.
//
// One node costs one call, not four: the seats are encoded back to back and go
// out as a single batch. Only the perspective rotates - identical weights serve
// every seat.
//
// Not safe for concurrent use; give each self-play worker its own. They can
// share a Session, or a Server's clients.
type Evaluator struct {
	runner Runner
	stride int
	in     []float32
	seats  [constrictor.MaxSnakes]int // batch row -> seat
}

func NewEvaluator(runner Runner, w, h int) *Evaluator {
	stride := constrictor.EncodeLen(w, h)
	return &Evaluator{
		runner: runner,
		stride: stride,
		in:     make([]float32, constrictor.MaxSnakes*stride),
	}
}

// Evaluate returns each living seat's prior over its own four moves and its
// expected placement. Dead seats keep zeros, which search ignores.
func (e *Evaluator) Evaluate(s *constrictor.State) (priors [constrictor.MaxSnakes][4]float32, values [constrictor.MaxSnakes]float64) {
	batch := 0
	for seat := 0; seat < s.N; seat++ {
		if !s.IsAlive(seat) {
			continue
		}
		s.Encode(seat, e.in[batch*e.stride:(batch+1)*e.stride])
		e.seats[batch] = seat
		batch++
	}
	if batch == 0 {
		return priors, values
	}

	policy, value, err := e.runner.Run(e.in[:batch*e.stride], batch)
	if err != nil {
		// A failed evaluation is a broken model or a broken build, not a game
		// state worth recovering from - and swallowing it would show up as a
		// bot that plays uniformly badly for reasons nobody can find.
		panic("nn: evaluating position: " + err.Error())
	}

	for row := 0; row < batch; row++ {
		seat := e.seats[row]
		priors[seat] = softmax(policy[row*MoveCount : (row+1)*MoveCount])
		values[seat] = float64(value[row])
	}
	return priors, values
}

// softmax turns the policy head's logits into a distribution. Shifted by the
// maximum first, which changes nothing mathematically and keeps Exp from
// overflowing on a confident network.
func softmax(logits []float32) [4]float32 {
	high := logits[0]
	for _, v := range logits[1:] {
		if v > high {
			high = v
		}
	}

	var out [4]float32
	total := float32(0)
	for i, v := range logits {
		out[i] = float32(math.Exp(float64(v - high)))
		total += out[i]
	}
	for i := range out {
		out[i] /= total
	}
	return out
}
