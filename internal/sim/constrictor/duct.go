package constrictor

// Decoupled UCT: one tree, but each node keeps separate statistics
// per seat over that seat's own four moves, and each seat selects independently.
//
// Leaves are scored by an Evaluator - a random rollout, or the network once it
// exists. Swapping the two changes nothing else here, which is the point.

import (
	"math"
	"math/rand"
)

// cPUCT weights exploration against the prior. Values are placements in
// [-1, 1]. A knob, not a law - pin it by measurement.
const cPUCT = 1.4

// Evaluator scores a leaf position. It returns, for every living seat, a prior
// over that seat's four moves and its expected placement, both from that seat's
// own perspective. Dead seats are ignored.
//
// The prior is what makes search affordable: unguided, DUCT has to spend
// simulations discovering that bodies are usually fatal, and measured against
// the Voronoi baseline that costs about 16x the budget. A prior concentrates
// search on plausible moves without asserting, as a hand-written mask would,
// that an occupied cell can never be entered.
// Only living seats need filling: a dead seat's placement is already decided,
// and search substitutes the real value rather than a predicted one.
type Evaluator interface {
	Evaluate(s *State) (priors [MaxSnakes][4]float32, values [MaxSnakes]float64)
}

// aliveMask is the Alive value of a position where nobody has died yet.
func aliveMask(seats int) uint8 { return 1<<seats - 1 }

// node holds one position's statistics, decoupled per seat: seat p's row is
// indexed by p's own move
type node struct {
	kids   map[[MaxSnakes]Move]*node // children by joint move, one entry per combination reached
	n      [MaxSnakes][4]int32       // [seat][move] simulations that selected it
	w      [MaxSnakes][4]float64     // [seat][move] placement scores summed over those; w/n is the mean
	prior  [MaxSnakes][4]float32     // [seat][move] the evaluator's opinion before any simulation
	visits int32                     // simulations through this node, the n of the PUCT bound
}

// Search configures one search. Sims and Eval are required.
type Search struct {
	Sims int
	Eval Evaluator
	Rnd  *rand.Rand

	// DirichletNoise mixes exploration noise into the root prior, as a weight
	// in [0, 1]. Self-play wants it so the tree does not keep replaying its own
	// opening; evaluation wants it off, to measure the policy as it is.
	DirichletNoise float64
}

// Result is the root statistics, which are both the move to play and the
// policy target to train on.
type Result struct {
	Visits [MaxSnakes][4]int32
}

// Run searches from s, which is not modified. All seats come out of one tree
// because DUCT builds all four sets of statistics anyway - a bot playing one
// seat just ignores the rest.
func (cfg Search) Run(s *State) Result {
	root := &node{kids: map[[MaxSnakes]Move]*node{}}
	root.prior, _ = cfg.Eval.Evaluate(s)
	if cfg.DirichletNoise > 0 {
		cfg.addNoise(s, root)
	}

	for i := 0; i < cfg.Sims; i++ {
		st := *s // no undo: State is 168 bytes of no pointers, copying beats unwinding
		cfg.simulate(root, &st)
	}
	return Result{Visits: root.n}
}

// Best returns the seat's most-visited move. Visit count rather than mean
// value: a high average off two visits is noise, while a high visit count means
// the bound kept coming back.
func (r Result) Best(seat int) Move {
	best, most := Up, int32(-1)
	for _, m := range AllMoves {
		if r.Visits[seat][m] > most {
			best, most = m, r.Visits[seat][m]
		}
	}
	return best
}

// Sample draws from the visit distribution, which is how self-play explores
// openings. temperature 0 is Best; 1 is proportional to visits.
func (r Result) Sample(seat int, temperature float64, rnd *rand.Rand) Move {
	if temperature <= 0 {
		return r.Best(seat)
	}

	var weights [4]float64
	total := 0.0
	for _, m := range AllMoves {
		if v := r.Visits[seat][m]; v > 0 {
			weights[m] = math.Pow(float64(v), 1/temperature)
			total += weights[m]
		}
	}
	if total == 0 {
		return Up // every move unvisited: the seat is walled in and dies regardless
	}

	draw := rnd.Float64() * total
	for _, m := range AllMoves {
		if draw -= weights[m]; draw < 0 {
			return m
		}
	}
	return Up
}

// Policy converts a seat's visit counts into the training target: the
// distribution search arrived at, which is a better move than the prior it
// started from. That gap is the whole learning signal.
func (r Result) Policy(seat int) [4]float32 {
	var out [4]float32
	total := float32(0)
	for _, m := range AllMoves {
		total += float32(r.Visits[seat][m])
	}
	if total == 0 {
		return out
	}
	for _, m := range AllMoves {
		out[m] = float32(r.Visits[seat][m]) / total
	}
	return out
}

// simulate descends to a leaf, evaluates it, and backs the placement scores up.
// s is advanced in place.
func (cfg Search) simulate(nd *node, s *State) [MaxSnakes]float64 {
	if s.Over() {
		return s.Placement()
	}

	alive := s.Alive
	joint := cfg.selectJoint(nd, s)
	kid, expanded := nd.kids[joint]
	s.Apply(joint)

	var scores [MaxSnakes]float64
	if expanded {
		scores = cfg.simulate(kid, s)
	} else {
		kid = &node{kids: map[[MaxSnakes]Move]*node{}}
		nd.kids[joint] = kid
		if s.Over() {
			scores = s.Placement()
		} else {
			kid.prior, scores = cfg.Eval.Evaluate(s)

			// A seat eliminated on the way here has a result, not a forecast:
			// every seat still alive outlasts it, so its placement is already
			// final and Placement computes it exactly. Leaving this to the
			// evaluator would back up 0 - a draw - for walking into a wall,
			// which is how a search talks itself into suicide.
			if final := s.Placement(); s.Alive != aliveMask(s.N) {
				for p := 0; p < s.N; p++ {
					if !s.IsAlive(p) {
						scores[p] = final[p]
					}
				}
			}
		}
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

// selectJoint picks each living seat's move independently by the PUCT bound
// over that seat's own statistics.
//
// Moves onto occupied cells are candidates like any other. Masking them would
// be wrong rather than merely wasteful: a body cell frees the instant its owner
// dies, so "occupied" is not the same as "fatal", and a mask would teach the
// tree that the cells opening up all around it in the late game do not exist.
// Only off-board moves are excluded, and only because a wall is the one thing
// on this board that can never stop being there.
func (cfg Search) selectJoint(nd *node, s *State) [MaxSnakes]Move {
	explore := cPUCT * math.Sqrt(float64(nd.visits)+1)

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
		offset := cfg.Rnd.Intn(count)
		best, bestScore := onBoard[offset], math.Inf(-1)
		for k := 0; k < count; k++ {
			m := onBoard[(k+offset)%count]

			visits := nd.n[p][m]
			q := 0.0 // unvisited sits at the neutral placement, neither promised nor written off
			if visits > 0 {
				q = nd.w[p][m] / float64(visits)
			}
			score := q + explore*float64(nd.prior[p][m])/(1+float64(visits))

			if score > bestScore {
				best, bestScore = m, score
			}
		}
		joint[p] = best
	}
	return joint
}

// addNoise mixes Dirichlet noise into the root prior so self-play does not
// replay one opening forever.
//
// ponytail: alpha is fixed at 1.0, where Dirichlet is just normalised
// exponentials and the sampler is one line. The plan wants alpha somewhere in
// 1.0-2.5; if tuning it proves worth the trouble, this needs a real gamma
// sampler (Marsaglia-Tsang) and nothing else.
func (cfg Search) addNoise(s *State, root *node) {
	for p := 0; p < s.N; p++ {
		if !s.IsAlive(p) {
			continue
		}
		var noise [4]float64
		total := 0.0
		for _, m := range AllMoves {
			noise[m] = -math.Log(1 - cfg.Rnd.Float64())
			total += noise[m]
		}
		for _, m := range AllMoves {
			mixed := (1-cfg.DirichletNoise)*float64(root.prior[p][m]) + cfg.DirichletNoise*noise[m]/total
			root.prior[p][m] = float32(mixed)
		}
	}
}

// RolloutEvaluator scores a leaf by playing it out at random, and offers no
// opinion on which move to try first. It is what the search uses before a
// network exists, and the control it is measured against afterwards.
//
// ponytail: measured against the Voronoi baseline, strength is flat from 50 to
// 800 sims and only moves at 3200 - the tree is fine, the leaf values are noise.
// Random play fills space badly, so a rollout barely predicts a space-filling
// game. Do not tune it; a trained network is the fix.
type RolloutEvaluator struct{ Rnd *rand.Rand }

func (e RolloutEvaluator) Evaluate(s *State) (priors [MaxSnakes][4]float32, values [MaxSnakes]float64) {
	for p := 0; p < s.N; p++ {
		for _, m := range AllMoves {
			priors[p][m] = 0.25
		}
	}

	st := *s
	for !st.Over() {
		var moves [MaxSnakes]Move
		for i := 0; i < st.N; i++ {
			if !st.IsAlive(i) {
				continue
			}
			// safeMoves is fine here in a way it is not in selection: this is a
			// rollout policy, a way to guess a leaf's value, and nothing it does
			// becomes a training target. Uniform-random moves would have every
			// snake walk into a wall in the opening and score nothing worth
			// backing up.
			if safe, n := safeMoves(&st, i); n > 0 {
				moves[i] = safe[e.Rnd.Intn(n)]
			}
		}
		st.Apply(moves)
	}
	return priors, st.Placement()
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
//
// Only for choosing how to *play* - rollouts, the Voronoi baseline, random test
// games. Never for pruning search, which is where that bias would end up in the
// policy target.
func safeMoves(s *State, i int) (moves [4]Move, n int) {
	for _, m := range AllMoves {
		if c, ok := s.step(s.Heads[i], m); ok && s.Cells[c] == Empty {
			moves[n] = m
			n++
		}
	}
	return moves, n
}
