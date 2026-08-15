package constrictor

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/stretchr/testify/require"

	"BattlesnakeReptarium/internal/model"
)

const (
	boardSize  = 11
	numSnakes  = 4
	diffGames  = 10000
	shortGames = 200
)

// TestMovesMatchDirections pins Move to its only reason for existing: an index
// into model.AllDirections. If the two lists drift apart the fast model plays a
// different game than the bot reports, and nothing else would catch it.
func TestMovesMatchDirections(t *testing.T) {
	require.Len(t, model.AllDirections, len(AllMoves))
	for i, m := range AllMoves {
		require.EqualValues(t, i, m, "AllMoves must be in index order")
	}
	require.Equal(t, model.UP, Up.Direction())
	require.Equal(t, model.DOWN, Down.Direction())
	require.Equal(t, model.LEFT, Left.Direction())
	require.Equal(t, model.RIGHT, Right.Direction())
}

// TestDifferentialAgainstOfficialRules plays
// random constrictor games in both the official pipeline and the fast model,
// asserting the boards stay identical every turn.
func TestDifferentialAgainstOfficialRules(t *testing.T) {
	games := diffGames
	if testing.Short() {
		games = shortGames
	}

	ruleset := rules.NewRulesetBuilder().NamedRuleset(rules.GameTypeConstrictor)
	rnd := rand.New(rand.NewSource(1))

	for g := 0; g < games; g++ {
		// Half the games start where the ladder starts them; the rest start
		// anywhere legal, purely to push the transition function through
		// positions fixed placement is too tidy to reach.
		seats := 2 + g%(numSnakes-1)
		bs, fast := newGame(t, rnd, seats, g%2 == 0)

		for turn := 0; !fast.Over(); turn++ {
			moves := pickMoves(fast, rnd)
			fast.Apply(moves)

			snakeMoves := make([]rules.SnakeMove, 0, seats)
			for i := range bs.Snakes {
				if bs.Snakes[i].EliminatedCause == rules.NotEliminated {
					snakeMoves = append(snakeMoves, rules.SnakeMove{ID: bs.Snakes[i].ID, Move: string(moves[i].Direction())})
				}
			}
			_, next, err := ruleset.Execute(bs, snakeMoves)
			require.NoError(t, err, "game %d turn %d", g, turn)
			bs = next
			bs.Turn++

			requireSameBoard(t, bs, fast, g, turn)
		}

		// Official game-over is <=1 snake left, same as fast.Over.
		over, _, err := ruleset.Execute(bs, nil)
		require.NoError(t, err)
		require.True(t, over, "game %d: fast model ended a game the official rules did not", g)
	}
}

// TestStartsMatchOfficialPlacement pins Starts to the two configurations the
// official fixed placement produces: all four corners, or all four cardinals.
func TestStartsMatchOfficialPlacement(t *testing.T) {
	official := map[string]bool{}
	for i := 0; i < 200; i++ {
		bs := rules.NewBoardState(boardSize, boardSize)
		require.NoError(t, rules.PlaceSnakesFixed(rules.NewSeedRand(int64(i)), bs, snakeIDs(numSnakes)))
		official[configKey(headCells(bs))] = true
	}

	mine := map[string]bool{}
	rnd := rand.New(rand.NewSource(4))
	for i := 0; i < 200; i++ {
		mine[configKey(Starts(boardSize, boardSize, numSnakes, rnd))] = true
	}

	require.Len(t, official, 2, "official placement should offer exactly two four-snake configurations")
	require.Equal(t, official, mine)
}

func TestThroughputGate(t *testing.T) {
	if testing.Short() {
		t.Skip("throughput gate is slow")
	}
	turns := 0
	rnd := rand.New(rand.NewSource(2))
	start := time.Now()
	for time.Since(start) < time.Second {
		for g := 0; g < 100; g++ {
			s := New(boardSize, boardSize, Starts(boardSize, boardSize, numSnakes, rnd))
			for !s.Over() {
				s.Apply(pickMoves(s, rnd))
				turns++
			}
		}
	}
	rate := float64(turns) / time.Since(start).Seconds()
	t.Logf("%.0f simulated turns/second", rate)
	require.Greater(t, rate, 200_000.0, "Phase 0 gate: >=200k simulated turns/second/core")
}

func TestPlacementScores(t *testing.T) {
	s := New(boardSize, boardSize, []uint8{0, 1, 2, 3})

	requirePlacement := func(died [MaxSnakes]int16, want ...float64) {
		t.Helper()
		s.Died = died
		got := s.Placement()
		require.InDeltaSlice(t, want, got[:], 1e-9)
	}

	requirePlacement([MaxSnakes]int16{-1, 9, 5, 2}, 1, 1.0/3, -1.0/3, -1) // 1st, 2nd, 3rd, 4th
	tied := (1.0/3 - 1.0/3 - 1) / 3
	requirePlacement([MaxSnakes]int16{-1, 4, 4, 4}, 1, tied, tied, tied) // winner, three-way tie for 2nd
	requirePlacement([MaxSnakes]int16{7, 7, 7, 7}, 0, 0, 0, 0)           // everyone out at once

	// Fewer seats than MaxSnakes: the unused entries are still -1 and must not
	// be scored as survivors sharing first place.
	duel := New(boardSize, boardSize, []uint8{0, 1})
	duel.Died = [MaxSnakes]int16{-1, 6, -1, -1}
	got := duel.Placement()
	require.InDeltaSlice(t, []float64{1, -1}, got[:2], 1e-9)
}

func BenchmarkApply(b *testing.B) {
	rnd := rand.New(rand.NewSource(3))
	s := New(boardSize, boardSize, Starts(boardSize, boardSize, numSnakes, rnd))
	for i := 0; i < b.N; i++ {
		if s.Over() {
			s = New(boardSize, boardSize, Starts(boardSize, boardSize, numSnakes, rnd))
		}
		s.Apply(pickMoves(s, rnd))
	}
}

// newGame builds the same starting position in both representations, taking the
// fast model's start cells from the official board so the two cannot disagree
// before a single move is played.
func newGame(t *testing.T, rnd *rand.Rand, seats int, fixedPlacement bool) (*rules.BoardState, *State) {
	t.Helper()
	bs := rules.NewBoardState(boardSize, boardSize)
	place := rules.PlaceSnakesRandomly
	if fixedPlacement {
		place = rules.PlaceSnakesFixed
	}
	require.NoError(t, place(rules.NewSeedRand(rnd.Int63()), bs, snakeIDs(seats)))
	return bs, New(boardSize, boardSize, headCells(bs))
}

func snakeIDs(n int) []string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = string(rune('a' + i))
	}
	return ids
}

func headCells(bs *rules.BoardState) []uint8 {
	cells := make([]uint8, len(bs.Snakes))
	for i, snake := range bs.Snakes {
		cells[i] = uint8(snake.Body[0].Y*bs.Width + snake.Body[0].X)
	}
	return cells
}

// configKey identifies a start configuration by its cells, ignoring which snake
// got which - placement shuffles that.
func configKey(cells []uint8) string {
	sorted := slices.Clone(cells)
	slices.Sort(sorted)
	return fmt.Sprint(sorted)
}

// pickMoves mostly plays moves into cells the fast model believes are free, so
// games run deep enough to exercise late-game collisions, and sometimes plays
// pure noise so a wrongly-rejected move still gets tried.
func pickMoves(s *State, rnd *rand.Rand) [MaxSnakes]Move {
	var moves [MaxSnakes]Move
	for i := 0; i < s.N; i++ {
		if !s.IsAlive(i) {
			continue
		}
		moves[i] = Move(rnd.Intn(4))
		if rnd.Intn(5) == 0 {
			continue
		}
		var free [4]Move
		n := 0
		for _, m := range AllMoves {
			if c, ok := s.step(s.Heads[i], m); ok && s.Cells[c] == Empty {
				free[n] = m
				n++
			}
		}
		if n > 0 {
			moves[i] = free[rnd.Intn(n)]
		}
	}
	return moves
}

func requireSameBoard(t *testing.T, bs *rules.BoardState, s *State, game, turn int) {
	t.Helper()

	var cells [MaxCells]uint8
	for i := range cells {
		cells[i] = Empty
	}
	for i, snake := range bs.Snakes {
		alive := snake.EliminatedCause == rules.NotEliminated
		require.Equal(t, alive, s.IsAlive(i),
			"game %d turn %d: snake %d alive mismatch (official cause %q)", game, turn, i, snake.EliminatedCause)
		if !alive {
			continue
		}
		require.Equal(t, uint8(snake.Body[0].Y*s.W+snake.Body[0].X), s.Heads[i],
			"game %d turn %d: snake %d head mismatch", game, turn, i)
		// Constrictor grows every snake every turn except the first (the start
		// body is three copies of one point, so the first move only unstacks it).
		require.Equal(t, max(3, bs.Turn+2), len(snake.Body),
			"game %d turn %d: snake %d has a length no other snake can have", game, turn, i)
		for _, p := range snake.Body {
			cells[p.Y*s.W+p.X] = uint8(i)
		}
	}
	require.Equal(t, cells, s.Cells, "game %d turn %d: occupancy mismatch", game, turn)
	require.Equal(t, bs.Turn, s.Turn, "game %d: turn counter mismatch", game)
}
