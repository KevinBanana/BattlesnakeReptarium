package selfplay

import "BattlesnakeReptarium/internal/sim/constrictor"

// VoronoiSearcher is the fixed yardstick as a Searcher, so a network can be
// measured against something that never changes. Self-play rating is purely
// relative and will climb happily inside a delusion; this is how that gets
// caught.
//
// It returns no policy, which is why evaluation against it records no training
// samples - and should not, since these are not the network's own moves.
type VoronoiSearcher struct{}

func (VoronoiSearcher) Search(s *constrictor.State) (moves [constrictor.MaxSnakes]constrictor.Move, policy [constrictor.MaxSnakes][4]float32) {
	for seat := 0; seat < s.N; seat++ {
		if s.IsAlive(seat) {
			moves[seat] = constrictor.Voronoi(s, seat)
		}
	}
	return moves, policy
}
