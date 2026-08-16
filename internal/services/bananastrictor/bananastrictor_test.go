package bananastrictor

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/sim/constrictor"
)

const (
	testModel = "../../../training/testdata/parity.onnx"
	testBoard = 11
)

// snake builds a snake whose body runs head-first through the given coords.
func snake(id string, coords ...model.Coord) model.Snake {
	return model.Snake{ID: id, Body: coords, Head: coords[0], Length: len(coords)}
}

func TestToState(t *testing.T) {
	me := snake("me", model.Coord{X: 1, Y: 1}, model.Coord{X: 1, Y: 0})
	them := snake("them", model.Coord{X: 9, Y: 9}, model.Coord{X: 9, Y: 10})
	board := model.Board{Width: testBoard, Height: testBoard, Snakes: []model.Snake{them, me}}

	state, seat, err := toState(board, me, 7, testBoard)
	require.NoError(t, err)

	// Seat 0 is always us, whatever order the board listed the snakes in.
	require.Equal(t, 0, seat)
	require.Equal(t, 2, state.N)
	require.Equal(t, 7, state.Turn)
	require.Equal(t, uint8(0b11), state.Alive)

	// Cell index is y*W+x, and the head is body[0].
	require.EqualValues(t, 1*testBoard+1, state.Heads[0])
	require.EqualValues(t, 9*testBoard+9, state.Heads[1])
	require.EqualValues(t, 0, state.Cells[1*testBoard+1], "our head")
	require.EqualValues(t, 0, state.Cells[0*testBoard+1], "our tail")
	require.EqualValues(t, 1, state.Cells[9*testBoard+9], "their head")
	require.EqualValues(t, 1, state.Cells[10*testBoard+9], "their tail")
	require.EqualValues(t, constrictor.Empty, state.Cells[5*testBoard+5], "empty middle")

	// The encoding must agree about which snake is whose - this is the
	// conversion's whole job, and a swap here would serve backwards moves.
	planes := make([]float32, constrictor.EncodeLen(testBoard, testBoard))
	state.Encode(seat, planes)
	cells := testBoard * testBoard
	require.Equal(t, float32(1), planes[constrictor.PlaneOwnHead*cells+1*testBoard+1])
	require.Equal(t, float32(1), planes[constrictor.PlaneEnemyHeads*cells+9*testBoard+9])
	require.Equal(t, float32(0), planes[constrictor.PlaneOwnBody*cells+9*testBoard+9])
}

func TestToStateRejects(t *testing.T) {
	me := snake("me", model.Coord{X: 1, Y: 1})

	t.Run("wrong board size", func(t *testing.T) {
		_, _, err := toState(model.Board{Width: 7, Height: 7, Snakes: []model.Snake{me}}, me, 0, testBoard)
		require.ErrorContains(t, err, "network expects")
	})

	t.Run("too many snakes", func(t *testing.T) {
		board := model.Board{Width: testBoard, Height: testBoard, Snakes: []model.Snake{
			me, snake("b", model.Coord{X: 2, Y: 2}), snake("c", model.Coord{X: 3, Y: 3}),
			snake("d", model.Coord{X: 4, Y: 4}), snake("e", model.Coord{X: 5, Y: 5}),
		}}
		_, _, err := toState(board, me, 0, testBoard)
		require.ErrorContains(t, err, "at most")
	})

	t.Run("self absent", func(t *testing.T) {
		board := model.Board{Width: testBoard, Height: testBoard, Snakes: []model.Snake{snake("other", model.Coord{X: 2, Y: 2})}}
		_, _, err := toState(board, me, 0, testBoard)
		require.ErrorContains(t, err, "not on the board")
	})
}

// TestCalculateMove runs the whole path a real request takes, from a board to a
// direction, with the network in the loop.
func TestCalculateMove(t *testing.T) {
	if _, err := os.Stat(testModel); err != nil {
		t.Skipf("no model at %s: run training/parity.py", testModel)
	}

	svc, err := open(testModel, 16, testBoard)
	require.NoError(t, err)
	defer svc.Close()

	require.Equal(t, []string{model.GamemodeConstrictor}, svc.Gamemodes())
	require.NotEmpty(t, svc.Customizations().Color)

	// Cornered on purpose: only up and right are even on the board, so a bot
	// that returns something else is not reading the position at all.
	me := snake("me", model.Coord{X: 0, Y: 0})
	them := snake("them", model.Coord{X: 10, Y: 10})
	board := model.Board{Width: testBoard, Height: testBoard, Snakes: []model.Snake{me, them}}

	action, err := svc.CalculateMove(context.Background(), model.Game{}, 0, board, me)
	require.NoError(t, err)
	require.Contains(t, []model.Direction{model.UP, model.RIGHT}, action.Move)
}

// TestWithoutWeights pins what a checkout with no trained network does: it
// still builds a bot, and that bot refuses moves with a message naming the file
// it wanted. The alternative - a nil Service in the bots map - would panic on
// the first request instead.
func TestWithoutWeights(t *testing.T) {
	_, err := open("does-not-exist.onnx", Sims, Board)
	require.ErrorContains(t, err, "does-not-exist.onnx")

	var svc Service // what New returns when the weights will not load
	require.NoError(t, svc.Close())

	_, err = svc.CalculateMove(context.Background(), model.Game{}, 0, model.Board{
		Width: Board, Height: Board,
		Snakes: []model.Snake{snake("me", model.Coord{X: 0, Y: 0})},
	}, snake("me", model.Coord{X: 0, Y: 0}))
	require.ErrorContains(t, err, Model)
}
