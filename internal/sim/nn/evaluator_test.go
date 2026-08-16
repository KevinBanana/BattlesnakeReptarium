package nn

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"BattlesnakeReptarium/internal/sim/constrictor"
)

// TestEvaluatorDrivesSearch runs a whole game with the network in the search
// loop. The weights are untrained, so this asserts nothing about strength -
// only that positions round-trip through encode, ONNX Runtime and softmax, and
// that search plays a legal game to a real finish.
//
// The turn count is also the regression for a bug the constrictor package could
// not reproduce with a stub: a seat eliminated on the way to a leaf had the
// evaluator's zero backed up instead of its real placement, so dying scored
// better than an untrained network's opinion of surviving. Games ended at turn
// 3; they now run to 16. It takes a real network to show, because the trap is a
// value that is confidently wrong rather than merely absent.
func TestEvaluatorDrivesSearch(t *testing.T) {
	const board = 11

	session := open(t, board)
	defer session.Close()

	rnd := rand.New(rand.NewSource(31))
	search := constrictor.Search{Sims: 32, Eval: NewEvaluator(session, board, board), Rnd: rnd}

	s := constrictor.New(board, board, constrictor.Starts(board, board, 4, rnd))
	for !s.Over() {
		result := search.Run(s)

		var moves [constrictor.MaxSnakes]constrictor.Move
		for seat := 0; seat < s.N; seat++ {
			if !s.IsAlive(seat) {
				continue
			}
			policy := result.Policy(seat)
			var total float32
			for _, p := range policy {
				total += p
			}
			require.InDelta(t, 1.0, total, 1e-5, "seat %d's visit distribution should be a distribution", seat)
			moves[seat] = result.Best(seat)
		}
		s.Apply(moves)
	}

	require.Greater(t, s.Turn, 8, "a game this short means seats are walking into walls on purpose")
	t.Logf("game ended at turn %d, placements %v", s.Turn, s.Placement())
}

// TestEvaluatorIgnoresDeadSeats pins the contract search relies on: a dead seat
// contributes no row to the batch and comes back as zeros.
func TestEvaluatorIgnoresDeadSeats(t *testing.T) {
	const board = 11

	session := open(t, board)
	defer session.Close()

	s := constrictor.New(board, board, []uint8{0, 5, 60})
	s.Apply([constrictor.MaxSnakes]constrictor.Move{constrictor.Down, constrictor.Up, constrictor.Up})
	require.False(t, s.IsAlive(0), "seat 0 should have walked off the bottom edge")

	priors, values := NewEvaluator(session, board, board).Evaluate(s)
	require.Equal(t, [4]float32{}, priors[0])
	require.Zero(t, values[0])

	for seat := 1; seat < s.N; seat++ {
		var total float32
		for _, p := range priors[seat] {
			total += p
		}
		require.InDelta(t, 1.0, total, 1e-5, "seat %d's prior should be a distribution", seat)
	}
}
