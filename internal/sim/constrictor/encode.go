package constrictor

// encode.go turns a position into network input. Ego-centric: the same weights
// serve every seat, and only the perspective rotates.
//
// Enemies are merged rather than given a plane each. Seat labels are arbitrary,
// so per-enemy planes make the network answer differently for two positions
// that are the same game - an encoding artifact that would then need permutation
// augmentation to wash out. Merging is permutation invariant by construction,
// and constrictor deletes two of the three things identity buys: bodies never
// recede, so there is no tail timing to attribute to an owner, and every living
// snake has length 3+turn, so every head-to-head resolves the same way whoever
// it is with.
//
// The third is real and is not deleted: identity is the partition, and the
// partition is which cells vanish together when a snake dies. Merged, the
// network cannot say that those twelve cells over there are one snake about to
// be killed. Three things keep it narrow:
//
//   - Search never has to predict it. Moves go through the real simulator, which
//     clears a dead snake's cells, so the position after the death is encoded
//     exactly. Only the leaf value at the node before it is guessing.
//   - How many is known even when which is not: every snake is length 3+turn,
//     and turn is a plane.
//   - A body is a connected path from its start cell to its head, so wherever
//     two enemy bodies do not touch, the heads plane partitions them anyway.
//
// It bites hardest on "move next to that head and bet on the trade opening the
// corridor" - which is a head-to-head standoff, where DUCT is separately weak.
// Per-enemy planes remain deferred; see the plan for what would trigger them.
//
// Whatever this file does, training/encode.py must do identically - it is the
// one piece of logic that exists in both languages, and the parity gate is what
// keeps them honest.

// Input plane layout. Dead snakes stop contributing cells, which encodes
// elimination for free, and an absent seat is indistinguishable from a dead one
// - which is what makes a 2- or 3-snake game work without special casing.
const (
	PlaneOwnBody = iota
	PlaneOwnHead
	PlaneEnemyBodies
	PlaneEnemyHeads
	PlaneTurn
	Planes
)

// EncodeLen is the length of the input for one board size.
func EncodeLen(w, h int) int { return Planes * w * h }

// Encode writes seat ego's view of the position into out, which must be
// EncodeLen(s.W, s.H) long. Planes are laid out one after another, each in the
// same y*W+x order as State.Cells, matching the NCHW input of the exported
// model.
func (s *State) Encode(ego int, out []float32) {
	for i := range out {
		out[i] = 0
	}
	cells := s.W * s.H

	for c := 0; c < cells; c++ {
		owner := s.Cells[c]
		switch {
		case owner == Empty:
		case int(owner) == ego:
			out[PlaneOwnBody*cells+c] = 1
		default:
			out[PlaneEnemyBodies*cells+c] = 1
		}
	}

	// Heads come from State.Heads rather than from Cells, which cannot tell a
	// head from the rest of a body. A dead seat's Heads entry is left stale by
	// Apply, so aliveness is checked here rather than trusted.
	for i := 0; i < s.N; i++ {
		if !s.IsAlive(i) {
			continue
		}
		plane := PlaneEnemyHeads
		if i == ego {
			plane = PlaneOwnHead
		}
		out[plane*cells+int(s.Heads[i])] = 1
	}

	// Turn scaled by the longest game the board can hold: every living snake is
	// length 3+turn, so two survivors on w*h cells run out by turn w*h/2 - 3.
	// The scale is a normalisation, not a meaning - the network learns what the
	// number is for.
	turn := float32(s.Turn) / (float32(cells)/2 - 3)
	for c := 0; c < cells; c++ {
		out[PlaneTurn*cells+c] = turn
	}
}
