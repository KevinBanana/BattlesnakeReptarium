package nn

import (
	"fmt"
	"math/rand"
	"testing"

	"BattlesnakeReptarium/internal/sim/constrictor"
)

// BenchmarkRun times one network call at the size search actually makes: a
// batch of 4, one row per seat. Self-play spends nearly all of its time here -
// one call per simulation, hundreds of simulations per move - so this number
// times sims times turns is the cost of a game.
// Batch sizes tell overhead from arithmetic: flat means the cost is per-call
// and worth hoisting, linear means the network is simply doing the work.
func BenchmarkRun(b *testing.B) {
	const board = 11

	session := openForBench(b, board)
	defer session.Close()

	for _, batch := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("batch-%d", batch), func(b *testing.B) {
			in := make([]float32, batch*constrictor.EncodeLen(board, board))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, _, err := session.Run(in, batch); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkEvaluate adds the encoding and softmax around it, which is what
// search actually calls.
func BenchmarkEvaluate(b *testing.B) {
	const board = 11

	session := openForBench(b, board)
	defer session.Close()

	evaluator := NewEvaluator(session, board, board)
	rnd := rand.New(rand.NewSource(1))
	s := constrictor.New(board, board, constrictor.Starts(board, board, 4, rnd))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.Evaluate(s)
	}
}

// BenchmarkSearch is one move at self-play settings.
func BenchmarkSearch(b *testing.B) {
	const board = 11

	session := openForBench(b, board)
	defer session.Close()

	rnd := rand.New(rand.NewSource(1))
	search := constrictor.Search{Sims: 200, Eval: NewEvaluator(session, board, board), Rnd: rnd}
	s := constrictor.New(board, board, constrictor.Starts(board, board, 4, rnd))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		search.Run(s)
	}
}
