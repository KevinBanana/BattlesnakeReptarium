package constrictor

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// 200 games puts the standard error near 0.05, so a positive result is a
	// result rather than seat luck.
	matchGames      = 200
	shortMatchGames = 8
	searchSims      = 3200
)

// TestSearchBeatsVoronoi is the plan's first real checkpoint: DUCT with random
// rollouts and no network at all should already outplay the space-filling
// baseline. If it does not, the simulator or the search is wrong, and that is
// worth finding out before any GPU is involved.
func TestSearchBeatsVoronoi(t *testing.T) {
	if testing.Short() {
		t.Skip("strength measurement is slow")
	}

	rnd := rand.New(rand.NewSource(11))
	search := Search{Sims: searchSims, Eval: RolloutEvaluator{Rnd: rnd}, Rnd: rnd}
	avg := match(rnd, games(), func(s *State, seat int) Move {
		return search.Run(s).Best(seat)
	})

	t.Logf("DUCT %d sims vs 3x Voronoi: average placement %+.3f", searchSims, avg)
	require.Greater(t, avg, 0.1, "search should beat the yardstick clearly, not just clear a chance baseline of 0")
}

// TestVoronoiBeatsRandom ensures Voronoi is better than random selection
func TestVoronoiBeatsRandom(t *testing.T) {
	rnd := rand.New(rand.NewSource(12))
	avg := match(rnd, games(), func(s *State, seat int) Move {
		safe, n := safeMoves(s, seat)
		if n == 0 {
			return Up
		}
		return safe[rnd.Intn(n)]
	})

	t.Logf("random-safe vs 3x Voronoi: average placement %+.3f", avg)
	require.Less(t, avg, 0.0)
}

// match plays out games with challenger at one seat and Voronoi at the other
// three, rotating the challenger through every start position, and returns its
// average placement
func match(rnd *rand.Rand, n int, challenger func(s *State, seat int) Move) float64 {
	total := 0.0
	for g := 0; g < n; g++ {
		seat := g % numSnakes
		s := New(boardSize, boardSize, Starts(boardSize, boardSize, numSnakes, rnd))

		for !s.Over() {
			var moves [MaxSnakes]Move
			for i := 0; i < s.N; i++ {
				switch {
				case !s.IsAlive(i):
				case i == seat:
					moves[i] = challenger(s, i)
				default:
					moves[i] = Voronoi(s, i)
				}
			}
			s.Apply(moves)
		}
		total += s.Placement()[seat]
	}
	return total / float64(n)
}

func games() int {
	if testing.Short() {
		return shortMatchGames
	}
	return matchGames
}
