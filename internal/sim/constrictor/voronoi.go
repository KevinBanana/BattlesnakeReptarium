package constrictor

// contested marks a cell two snakes reach on the same step, so it counts for
// neither.
const contested = 254

// Voronoi picks the move reaching the most cells first, by multi-source flood
// fill from every living head at once.
func Voronoi(s *State, me int) Move {
	safe, n := safeMoves(s, me)
	if n == 0 {
		return Up // doomed either way
	}

	best, bestScore := safe[0], -1
	for k := 0; k < n; k++ {
		c, _ := s.step(s.Heads[me], safe[k])
		score := territory(s, me, c)
		if headAdjacent(s, me, c) {
			score -= MaxCells
		}
		if score > bestScore {
			best, bestScore = safe[k], score
		}
	}
	return best
}

// territory counts the cells my head reaches strictly before any enemy head
func territory(s *State, me int, myHead uint8) int {
	var dist [MaxCells]int16
	var owner [MaxCells]uint8
	for i := range dist {
		dist[i] = -1
	}

	var queue [MaxCells]uint8
	n := 0
	seed := func(c uint8, o int) {
		dist[c], owner[c] = 0, uint8(o)
		queue[n] = c
		n++
	}
	seed(myHead, me)
	for i := 0; i < s.N; i++ {
		if i != me && s.IsAlive(i) {
			seed(s.Heads[i], i)
		}
	}

	for r := 0; r < n; r++ {
		c := queue[r]
		for _, m := range AllMoves {
			nc, ok := s.step(c, m)
			if !ok || s.Cells[nc] != Empty {
				continue
			}
			switch {
			case dist[nc] == -1:
				dist[nc], owner[nc] = dist[c]+1, owner[c]
				queue[n] = nc
				n++
			case dist[nc] == dist[c]+1 && owner[nc] != owner[c]:
				owner[nc] = contested
			}
		}
	}

	// Counted after the fill rather than as cells are popped: a cell can be
	// marked contested by a rival that is still sitting in the queue behind it.
	count := 0
	for c := 0; c < s.W*s.H; c++ {
		if dist[c] != -1 && owner[c] == uint8(me) {
			count++
		}
	}
	return count
}

// headAdjacent reports whether a living enemy could move onto c this turn.
func headAdjacent(s *State, me int, c uint8) bool {
	for i := 0; i < s.N; i++ {
		if i == me || !s.IsAlive(i) {
			continue
		}
		for _, m := range AllMoves {
			if nc, ok := s.step(s.Heads[i], m); ok && nc == c {
				return true
			}
		}
	}
	return false
}
