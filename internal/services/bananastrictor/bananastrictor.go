// bananastrictor plays with search rather than the policy head alone. The raw network is
// meaningfully weaker than the network plus a tree, and the tree is where the
// exact rules live.
package bananastrictor

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/sim/constrictor"
	"BattlesnakeReptarium/internal/sim/nn"
)

const (
	Name  = "bananastrictor"
	Model = "weights/bananastrictor.onnx"

	// Budget is how long a move may take, and the real limit on search. The
	// ladder allows ~500ms including the network round trip and the engine's own
	// overhead, so this leaves room for both.
	Budget = 300 * time.Millisecond
	Sims   = 2000

	// Board is the size the network was trained on
	Board = 11
)

type Service struct {
	session     *nn.Session
	sims, board int
	budget      time.Duration

	// Evaluators hold a scratch buffer and are not safe for concurrent use, so
	// each move borrows one. The ORT session underneath is shared and is safe.
	evaluators sync.Pool
}

func New() *Service {
	svc, err := open(Model, Sims, Budget, Board)
	if err != nil {
		slog.Error("bananastrictor has no weights and will refuse moves", "model", Model, "err", err)
		return &Service{}
	}

	slog.Info("bananastrictor loaded", "model", Model, "sims", Sims, "budget", Budget, "board", Board)
	return svc
}

// open is New with the parameters visible, so tests can point at a model of
// their own without a file at the deployment path.
func open(path string, sims int, budget time.Duration, board int) (*Service, error) {
	session, err := nn.Open(path, constrictor.Planes, board, board)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", path, err)
	}

	svc := &Service{session: session, sims: sims, budget: budget, board: board}
	svc.evaluators.New = func() any { return nn.NewEvaluator(session, board, board) }
	return svc, nil
}

func (svc *Service) Customizations() model.SnakeCustomizations {
	return model.SnakeCustomizations{
		Head:  "all-seeing",
		Tail:  "block-bum",
		Color: "#f4d35e",
	}
}

func (svc *Service) Gamemodes() []string {
	return []string{model.GamemodeConstrictor}
}

func (svc *Service) CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, selfSnake model.Snake) (*model.SnakeAction, error) {
	if svc.session == nil {
		return nil, fmt.Errorf("bananastrictor has no weights loaded from %s", Model)
	}

	state, seat, err := toState(board, selfSnake, turn, svc.board)
	if err != nil {
		return nil, err
	}

	evaluator := svc.evaluators.Get().(*nn.Evaluator)
	defer svc.evaluators.Put(evaluator)

	search := constrictor.Search{
		Sims:     svc.sims,
		Deadline: deadline(ctx, svc.budget),
		Eval:     evaluator,
		Rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	move := search.Run(state).Best(seat)

	return &model.SnakeAction{Move: move.Direction()}, nil
}

// deadline returns when search must stop: our own budget, or the caller's if
// it is sooner. A request that has already been waiting has less time left than
// we would otherwise assume.
func deadline(ctx context.Context, budget time.Duration) time.Time {
	ours := time.Now().Add(budget)
	if theirs, ok := ctx.Deadline(); ok && theirs.Before(ours) {
		return theirs
	}
	return ours
}

func (svc *Service) Close() error {
	if svc.session == nil {
		return nil
	}
	return svc.session.Close()
}

// toState re-expresses the API's board as the fast model's, and returns the
// seat the caller occupies.
func toState(board model.Board, selfSnake model.Snake, turn, want int) (*constrictor.State, int, error) {
	if board.Width != want || board.Height != want {
		return nil, 0, fmt.Errorf("board is %dx%d, network expects %dx%d", board.Width, board.Height, want, want)
	}
	if len(board.Snakes) > constrictor.MaxSnakes {
		return nil, 0, fmt.Errorf("%d snakes, network expects at most %d", len(board.Snakes), constrictor.MaxSnakes)
	}

	state := &constrictor.State{W: board.Width, H: board.Height, N: len(board.Snakes), Turn: turn}
	for i := range state.Cells {
		state.Cells[i] = constrictor.Empty
	}
	for i := range state.Died {
		state.Died[i] = -1
	}

	seats := make([]model.Snake, 0, len(board.Snakes))
	seats = append(seats, selfSnake)
	for _, snake := range board.Snakes {
		if snake.ID != selfSnake.ID {
			seats = append(seats, snake)
		}
	}
	if len(seats) != len(board.Snakes) {
		return nil, 0, fmt.Errorf("snake %q is not on the board", selfSnake.ID)
	}

	for seat, snake := range seats {
		if len(snake.Body) == 0 {
			return nil, 0, fmt.Errorf("snake %q has no body", snake.ID)
		}
		for _, part := range snake.Body {
			state.Cells[part.Y*board.Width+part.X] = uint8(seat)
		}
		state.Heads[seat] = uint8(snake.Body[0].Y*board.Width + snake.Body[0].X)
		state.Alive |= 1 << seat
	}

	return state, 0, nil
}
