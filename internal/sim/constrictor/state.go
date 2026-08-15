// Package constrictor is the fast forward model for constrictor games, used by
// search. Other gamemodes get their own package under internal/sim - the state
// a mode can ignore is exactly what makes its model fast, so there is nothing
// worth sharing between them beyond the setup helpers here.
//
// Constrictor tracks occupancy only: every living snake has the same length and
// bodies never recede, so a cell is either blocked or it is not - no tail
// timing, no health, no food. The only thing that ever frees a cell is a snake
// dying, and its whole body goes at once.
//
// State transitions are a hand-written mirror of the official constrictor
// pipeline in github.com/BattlesnakeOfficial/rules; state_test.go asserts they
// agree over random games.
package constrictor

import (
	"math/bits"
	"math/rand"

	"BattlesnakeReptarium/internal/model"
)

const (
	MaxSnakes = 4
	MaxCells  = 11 * 11

	// Empty marks a free cell in State.Cells.
	Empty = 255
)

// Move is model.Direction as an index. Same four directions, but search switches
// on them millions of times per iteration and the policy head emits four logits,
// so the hot path wants 0-3 rather than the API's strings.
//
// The values are positions in model.AllDirections, so converting is one lookup
// with nothing to keep in sync. Reordering either list breaks the other, which
// is what TestMovesMatchDirections is for.
type Move uint8

const (
	Up Move = iota
	Left
	Down
	Right
)

var AllMoves = [4]Move{Up, Left, Down, Right}

func (m Move) Direction() model.Direction { return model.AllDirections[m] }

// State is a complete constrictor position. Cell index is y*W + x.
type State struct {
	W, H int
	N    int // number of seats
	Turn int
	// Cells holds the bodies: snake i's body is every cell owning index i, with
	// no segment order because nothing here needs one. Bodies never recede, so a
	// body is just the cells that snake's head has visited.
	Cells [MaxCells]uint8  // owning snake index, or Empty
	Heads [MaxSnakes]uint8 // current head, the only body cell a move starts from
	Alive uint8            // bitmask of living snakes
	Died  [MaxSnakes]int16 // turn of elimination, -1 while alive
}

// New returns a turn-zero state with one snake per start cell. Start cells must
// be distinct. For a real game position use Starts to generate them; arbitrary
// cells are accepted so that a mid-game board can be loaded too.
func New(w, h int, starts []uint8) *State {
	s := &State{W: w, H: h, N: len(starts)}
	for i := range s.Cells {
		s.Cells[i] = Empty
	}
	for i, c := range starts {
		s.Cells[c] = uint8(i)
		s.Heads[i] = c
		s.Alive |= 1 << i
	}
	for i := range s.Died {
		s.Died[i] = -1
	}
	return s
}

// Starts returns start cells drawn the way the official fixed placement draws
// them: either the four corner points or the four cardinal edge points, never a
// mix for four or fewer snakes, shuffled within the group. Real games only ever
// begin in one of those two configurations, so neither does self-play.
func Starts(w, h, n int, rnd *rand.Rand) []uint8 {
	mn, md, mx := 1, (w-1)/2, w-2
	group := [4][2]int{{mn, mn}, {mn, mx}, {mx, mn}, {mx, mx}} // corners
	if rnd.Intn(2) == 0 {
		group = [4][2]int{{mn, md}, {md, mn}, {md, mx}, {mx, md}} // cardinals
	}
	rnd.Shuffle(len(group), func(i, j int) { group[i], group[j] = group[j], group[i] })

	starts := make([]uint8, n)
	for i := range starts {
		starts[i] = uint8(group[i][1]*w + group[i][0])
	}
	return starts
}

func (s *State) IsAlive(i int) bool { return s.Alive&(1<<i) != 0 }

// Over reports whether the game has ended. Checked before moves are applied,
// matching the game_over stage sitting first in the official pipeline.
func (s *State) Over() bool { return bits.OnesCount8(s.Alive) <= 1 }

// step returns the cell reached by moving from c, and whether it is on the board.
func (s *State) step(c uint8, m Move) (uint8, bool) {
	x, y := int(c)%s.W, int(c)/s.W
	switch m {
	case Up:
		y++
	case Down:
		y--
	case Left:
		x--
	case Right:
		x++
	}
	if x < 0 || x >= s.W || y < 0 || y >= s.H {
		return 0, false
	}
	return uint8(y*s.W + x), true
}

// Apply advances one turn in place. moves[i] is ignored for dead snakes.
func (s *State) Apply(moves [MaxSnakes]Move) {
	var newHead [MaxSnakes]uint8
	var oob, dying uint8
	for i := 0; i < s.N; i++ {
		if !s.IsAlive(i) {
			continue
		}
		c, ok := s.step(s.Heads[i], moves[i])
		if !ok {
			oob |= 1 << i
			continue
		}
		newHead[i] = c
	}

	// The official elimination stage removes out-of-bounds snakes first and then
	// skips them when resolving collisions, so a wall-dead snake's body does not
	// kill anyone this turn. survivors is exactly that "still a hazard" set.
	survivors := s.Alive &^ oob
	for i := 0; i < s.N; i++ {
		if survivors&(1<<i) == 0 {
			continue
		}
		// Own body included: Cells[newHead] == i is a self-collision.
		if owner := s.Cells[newHead[i]]; owner != Empty && survivors&(1<<owner) != 0 {
			dying |= 1 << i
			continue
		}
		for j := 0; j < s.N; j++ {
			// Every living snake is the same length, so nobody wins a head-to-head.
			if j != i && survivors&(1<<j) != 0 && newHead[j] == newHead[i] {
				dying |= 1 << i
				break
			}
		}
	}

	// Dead bodies leave the board. Clear them before writing new heads: a
	// survivor may have moved onto a cell freed by a snake that hit a wall.
	dead := oob | dying
	if dead != 0 {
		for c := 0; c < s.W*s.H; c++ {
			if o := s.Cells[c]; o != Empty && dead&(1<<o) != 0 {
				s.Cells[c] = Empty
			}
		}
	}
	for i := 0; i < s.N; i++ {
		if dead&(1<<i) != 0 {
			s.Died[i] = int16(s.Turn + 1)
			continue
		}
		if !s.IsAlive(i) {
			continue
		}
		s.Cells[newHead[i]] = uint8(i)
		s.Heads[i] = newHead[i]
	}
	s.Alive &^= dead
	s.Turn++
}

// Placement scores each seat by finishing position: 1st = +1, last = -1, evenly
// spaced between. Snakes eliminated on the same turn share the average of the
// positions they span. Chance baseline is 0.
func (s *State) Placement() [MaxSnakes]float64 {
	var out [MaxSnakes]float64
	if s.N < 2 {
		return out
	}
	for i := 0; i < s.N; i++ {
		better, tied := 0, 0
		for j := 0; j < s.N; j++ {
			switch {
			case s.outlasted(j, i):
				better++
			case s.Died[j] == s.Died[i]:
				tied++
			}
		}
		// Ranks better+1 .. better+tied, averaged.
		rank := float64(better) + float64(tied+1)/2
		out[i] = 1 - 2*(rank-1)/float64(s.N-1)
	}
	return out
}

// outlasted reports whether j finished ahead of i. -1 (still alive) beats every
// real elimination turn.
func (s *State) outlasted(j, i int) bool {
	if s.Died[j] == s.Died[i] {
		return false
	}
	return s.Died[j] == -1 || (s.Died[i] != -1 && s.Died[j] > s.Died[i])
}
