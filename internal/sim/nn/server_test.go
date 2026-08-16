package nn

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"BattlesnakeReptarium/internal/sim/constrictor"
)

// TestServerMatchesDirect is the property that makes the server safe to use:
// batching changes when a position is evaluated, never what comes back. A
// scatter bug - handing a worker somebody else's rows - would be invisible in
// self-play, showing up only as a network that never quite learns.
func TestServerMatchesDirect(t *testing.T) {
	const board = 11

	session := open(t, board)
	defer session.Close()

	// Positions of varying seat counts, so requests have varying row counts and
	// the offsets have something to get wrong.
	rnd := rand.New(rand.NewSource(7))
	var states []*constrictor.State
	for seats := 2; seats <= 4; seats++ {
		s := constrictor.New(board, board, constrictor.Starts(board, board, seats, rnd))
		for turn := 0; turn < 4 && !s.Over(); turn++ {
			states = append(states, &constrictor.State{W: s.W, H: s.H, N: s.N, Turn: s.Turn,
				Cells: s.Cells, Heads: s.Heads, Alive: s.Alive, Died: s.Died})
			var moves [constrictor.MaxSnakes]constrictor.Move
			for seat := 0; seat < s.N; seat++ {
				moves[seat] = constrictor.Move(rnd.Intn(4))
			}
			s.Apply(moves)
		}
	}
	require.NotEmpty(t, states)

	direct := NewEvaluator(session, board, board)
	wantPriors := make([][constrictor.MaxSnakes][4]float32, len(states))
	wantValues := make([][constrictor.MaxSnakes]float64, len(states))
	livingSeats := 0
	for i, s := range states {
		wantPriors[i], wantValues[i] = direct.Evaluate(s)
		for seat := 0; seat < s.N; seat++ {
			if s.IsAlive(seat) {
				livingSeats++
			}
		}
	}

	// Deliberately more workers than states and a small cap, so batches are
	// assembled from several requests and split again.
	const workers = 16
	server := NewServer(session, 8, 64, 200*time.Microsecond)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			evaluator := NewEvaluator(server.Client(), board, board)
			for i, s := range states {
				priors, values := evaluator.Evaluate(s)
				// assert, not require: require calls FailNow, which may only be
				// used from the goroutine running the test.
				assert.Equal(t, wantPriors[i], priors, "worker %d, state %d", worker, i)
				assert.Equal(t, wantValues[i], values, "worker %d, state %d", worker, i)
			}
		}(w)
	}
	wg.Wait()
	server.Close()

	calls, rows := server.Stats()
	require.Equal(t, int64(workers*livingSeats), rows, "every seat of every evaluation should have gone through")
	require.Less(t, calls, rows, "nothing was batched at all")
	t.Logf("%d positions in %d calls, average batch %.1f", rows, calls, float64(rows)/float64(calls))
}
