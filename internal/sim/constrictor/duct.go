package constrictor

// Decoupled UCT: one tree, but each node keeps separate statistics
// per seat over that seat's own four moves, and each seat selects independently.
//
// Leaves are evaluated by a random rollout to a terminal state.

import (
	"math"
	"math/rand"
)

// exploreC is the UCT exploration weight. Values are placements in [-1, 1].
const exploreC = 1.4

// node holds one position's statistics, decoupled per seat: seat p's row is
// indexed by p's own move
type node struct {
	kids   map[[MaxSnakes]Move]*node // children by joint move, one entry per combination reached
	n      [MaxSnakes][4]int32       // [seat][move] simulations that selected it
	w      [MaxSnakes][4]float64     // [seat][move] placement scores summed over those; w/n is the mean
	visits int32                     // simulations through this node, the n of the UCT bound
}

// Search runs sims simulations from s and returns every seat's most-visited
// move. s is not modified. All four seats come out of one tree.
func Search(s *State, sims int, rnd *rand.Rand) [MaxSnakes]Move {
	root := &node{kids: map[[MaxSnakes]Move]*node{}}
	for i := 0; i < sims; i++ {
		st := *s
		simulate(root, &st, rnd)
	}

	var out [MaxSnakes]Move
	for p := 0; p < s.N; p++ {
		best := int32(-1)
		for _, m := range AllMoves {
			if root.n[p][m] > best {
				best, out[p] = root.n[p][m], m
			}
		}
	}
	return out
}

// simulate descends to a leaf, rolls out from it, and backs the placement
// scores up. s is advanced in place.
func simulate(nd *node, s *State, rnd *rand.Rand) [MaxSnakes]float64 {
	if s.Over() {
		return s.Placement()
	}

	alive := s.Alive
	joint := nd.selectJoint(s, rnd)
	kid, expanded := nd.kids[joint]
	s.Apply(joint)

	var scores [MaxSnakes]float64
	if expanded {
		scores = simulate(kid, s, rnd)
	} else {
		nd.kids[joint] = &node{kids: map[[MaxSnakes]Move]*node{}}
		scores = rollout(s, rnd)
	}

	nd.visits++
	for p := 0; p < s.N; p++ {
		if alive&(1<<p) == 0 {
			continue
		}
		nd.n[p][joint[p]]++
		nd.w[p][joint[p]] += scores[p]
	}
	return scores
}

// selectJoint picks each living seat's move independently by the UCT bound over
// that seat's own statistics. Untried moves go first.
//
// Moves onto occupied cells are candidates like any other. Masking them would
// be wrong rather than merely wasteful: a body cell frees the instant its owner
// dies, so "occupied" is not the same as "fatal", and a mask would teach the
// tree that the cells opening up all around it in the late game do not exist.
// Only off-board moves are excluded, and only because a wall is the one thing
// on this board that can never stop being there.
func (nd *node) selectJoint(s *State, rnd *rand.Rand) [MaxSnakes]Move {
	logN := math.Log(float64(nd.visits) + 1)

	var joint [MaxSnakes]Move
	for p := 0; p < s.N; p++ {
		if !s.IsAlive(p) {
			continue
		}
		onBoard, count := onBoardMoves(s, p)
		if count == 0 {
			continue // walled in; the move it dies playing does not matter
		}

		// Each seat starts scanning at a different random move. Scanning in list
		// order instead makes every seat break its ties the same way, so a fresh
		// node spends its first simulations on the four joint moves where all
		// four snakes go the same direction - 4 of 256, and the least
		// representative 4.
		offset := rnd.Intn(count)
		best, bestScore := onBoard[offset], math.Inf(-1)
		for k := 0; k < count; k++ {
			m := onBoard[(k+offset)%count]
			visits := nd.n[p][m]
			if visits == 0 {
				best = m
				break
			}
			score := nd.w[p][m]/float64(visits) + exploreC*math.Sqrt(logN/float64(visits))
			if score > bestScore {
				best, bestScore = m, score
			}
		}
		joint[p] = best
	}
	return joint
}

// rollout plays random non-suicidal moves to the end and returns the placements.
// Uniform-random moves would have every snake walk into a wall in the opening
// and score nothing worth backing up.
//
// safeMoves is fine here in a way it is not in selection: this is a rollout
// policy, a way to guess a leaf's value, and nothing it does becomes a training
// target. Phase 1 deletes it for the value head anyway - do not tune it.
func rollout(s *State, rnd *rand.Rand) [MaxSnakes]float64 {
	for !s.Over() {
		var moves [MaxSnakes]Move
		for i := 0; i < s.N; i++ {
			if !s.IsAlive(i) {
				continue
			}
			if safe, n := safeMoves(s, i); n > 0 {
				moves[i] = safe[rnd.Intn(n)]
			}
		}
		s.Apply(moves)
	}
	return s.Placement()
}

// onBoardMoves returns the moves that keep seat i on the board. This is an
// exact legality test, not a heuristic - every excluded move is fatal in every
// continuation - so it is the only filter allowed to touch search selection,
// where the visit counts become a training target.
func onBoardMoves(s *State, i int) (moves [4]Move, n int) {
	for _, m := range AllMoves {
		if _, ok := s.step(s.Heads[i], m); ok {
			moves[n] = m
			n++
		}
	}
	return moves, n
}

// safeMoves returns the moves onto a cell nothing currently occupies. Unlike
// onBoardMoves this is a guess: an occupied cell frees up the moment its owner
// dies, so a survivable move is sometimes excluded.
func safeMoves(s *State, i int) (moves [4]Move, n int) {
	for _, m := range AllMoves {
		if c, ok := s.step(s.Heads[i], m); ok && s.Cells[c] == Empty {
			moves[n] = m
			n++
		}
	}
	return moves, n
}
