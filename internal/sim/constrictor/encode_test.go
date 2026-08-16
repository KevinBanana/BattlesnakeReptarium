package constrictor

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const parityFixture = "testdata/encode_parity.json"

func TestEncode(t *testing.T) {
	// Seat 0 bottom-left, seat 1 bottom-right, seat 2 top-left. 3x3 board, and
	// seat 2 walks off the top edge so the dead-seat cases are covered too.
	s := New(3, 3, []uint8{0, 2, 6})
	s.Apply([MaxSnakes]Move{Up, Up, Up})
	require.Equal(t, uint8(0b011), s.Alive, "seats 0 and 1 alive, seat 2 off the board")

	out := make([]float32, EncodeLen(3, 3))
	s.Encode(1, out)

	plane := func(i int) string {
		var b []byte
		for c := 0; c < 9; c++ {
			if out[i*9+c] != 0 {
				b = append(b, '#')
			} else {
				b = append(b, '.')
			}
		}
		return string(b)
	}

	// Cell order is y*W+x, so each string reads bottom row first.
	require.Equal(t, "..#"+"..#"+"...", plane(PlaneOwnBody), "seat 1's start cell and its new head")
	require.Equal(t, "..."+"..#"+"...", plane(PlaneOwnHead))
	require.Equal(t, "#.."+"#.."+"...", plane(PlaneEnemyBodies), "seat 0 only; seat 2's body left the board with it")
	require.Equal(t, "..."+"#.."+"...", plane(PlaneEnemyHeads), "one head, not two - dead seats have stale Heads entries")

	// turn 1 of a 3x3 board's 9/2-3 = 1.5 turn ceiling
	require.InDelta(t, 1/1.5, out[PlaneTurn*9], 1e-6)
}

// TestEncodeIgnoresEnemyIdentity pins the property the merged-enemy planes exist
// for: relabelling two enemy seats is the same game, so it must be the same
// input. Nothing else in the pipeline enforces it, and the whole reason there is
// no permutation augmentation is that this holds.
func TestEncodeIgnoresEnemyIdentity(t *testing.T) {
	rnd := rand.New(rand.NewSource(21))

	for g := 0; g < 200; g++ {
		s := New(boardSize, boardSize, Starts(boardSize, boardSize, numSnakes, rnd))
		for !s.Over() {
			s.Apply(pickMoves(s, rnd))

			want := make([]float32, EncodeLen(s.W, s.H))
			got := make([]float32, EncodeLen(s.W, s.H))
			s.Encode(0, want)
			swapSeats(s, 1, 2).Encode(0, got)
			require.Equal(t, want, got, "game %d turn %d: swapping seats 1 and 2 changed the input", g, s.Turn)
		}
	}
}

// TestWriteParityFixture regenerates the positions the Python encoder is checked
// against. Run it when the encoding changes, and commit the result:
//
//	go test ./internal/sim/constrictor -run WriteParityFixture
//	python training/parity.py
func TestWriteParityFixture(t *testing.T) {
	type sample struct {
		W, H, N int       `json:"-"`
		Width   int       `json:"w"`
		Height  int       `json:"h"`
		Seats   int       `json:"n"`
		Turn    int       `json:"turn"`
		Alive   uint8     `json:"alive"`
		Cells   []int     `json:"cells"` // int, not uint8: encoding/json base64s a []byte
		Heads   []int     `json:"heads"`
		Ego     int       `json:"ego"`
		Encoded []float32 `json:"encoded"`
	}

	rnd := rand.New(rand.NewSource(22))
	var samples []sample

	// Sampled across whole games so the fixture covers openings, mid-game, and
	// positions with seats already eliminated.
	for g := 0; g < 6; g++ {
		s := New(boardSize, boardSize, Starts(boardSize, boardSize, numSnakes, rnd))
		for turn := 0; !s.Over(); turn++ {
			if turn%4 == 0 {
				ego := rnd.Intn(s.N)
				enc := make([]float32, EncodeLen(s.W, s.H))
				s.Encode(ego, enc)
				samples = append(samples, sample{
					Width: s.W, Height: s.H, Seats: s.N, Turn: s.Turn, Alive: s.Alive,
					Cells: ints(s.Cells[:s.W*s.H]), Heads: ints(s.Heads[:]), Ego: ego, Encoded: enc,
				})
			}
			s.Apply(pickMoves(s, rnd))
		}
	}

	body, err := json.Marshal(map[string]any{"planes": Planes, "samples": samples})
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(parityFixture), 0o750))
	require.NoError(t, os.WriteFile(parityFixture, body, 0o600))
	t.Logf("%d samples written to %s", len(samples), parityFixture)
}

func ints(b []uint8) []int {
	out := make([]int, len(b))
	for i, v := range b {
		out[i] = int(v)
	}
	return out
}

// swapSeats returns s with two seats relabelled - the same position, described
// differently.
func swapSeats(s *State, a, b int) *State {
	out := *s
	for c := 0; c < out.W*out.H; c++ {
		switch out.Cells[c] {
		case uint8(a):
			out.Cells[c] = uint8(b)
		case uint8(b):
			out.Cells[c] = uint8(a)
		}
	}
	out.Heads[a], out.Heads[b] = out.Heads[b], out.Heads[a]
	out.Died[a], out.Died[b] = out.Died[b], out.Died[a]

	out.Alive = s.Alive &^ (1<<a | 1<<b)
	if s.IsAlive(a) {
		out.Alive |= 1 << b
	}
	if s.IsAlive(b) {
		out.Alive |= 1 << a
	}
	return &out
}
