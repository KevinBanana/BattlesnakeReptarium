// Package selfplay plays batches of games and reports what happened.
//
// Two jobs, because the loop needs two: self-play, where every seat is the same
// network and the games become training records, and evaluation, where one seat
// plays something else and only the placements matter.
package selfplay

import (
	"math/rand"
	"sync"

	"BattlesnakeReptarium/internal/sim/constrictor"
)

// Searcher decides a whole turn at once: every seat's move, and the visit
// distribution behind it.
//
// One call rather than one per seat, because DUCT builds all four seats'
// statistics from a single tree - asking seat by seat would cost four searches
// for the same answer. A caller that only wants one seat's move takes it and
// ignores the rest.
type Searcher interface {
	Search(s *constrictor.State) (moves [constrictor.MaxSnakes]constrictor.Move, policy [constrictor.MaxSnakes][4]float32)
}

// Batch is a set of games to play.
type Batch struct {
	Board, Seats int
	Games        int
	Workers      int
	Seed         int64

	// Progress is called after each finished game, from several goroutines.
	Progress func(done int)
}

// Result is what a batch produced.
type Result struct {
	Samples []constrictor.Sample
	Frames  []constrictor.State // the first game, kept for rendering
	Games   int

	// Placement sums the subject's placement over the batch; Ranks counts its
	// finishes, Ranks[0] being firsts. Under self-play the subject is seat 0
	// and both are of little interest.
	Placement float64
	Ranks     [constrictor.MaxSnakes]int
}

// Average placement over the batch. Chance is 0; +1 is winning every game.
func (r Result) Average() float64 {
	if r.Games == 0 {
		return 0
	}
	return r.Placement / float64(r.Games)
}

// SelfPlay runs games where every seat is the same searcher, and keeps the
// training records. searcher is called once per worker, so each gets its own.
func SelfPlay(b Batch, searcher func(worker int) Searcher) Result {
	return run(b, func(worker int) table {
		s := searcher(worker)
		return table{subject: s, opponent: s, record: true}
	})
}

// Evaluate seats the subject against three opponents, rotating it through every
// start position in equal numbers, and keeps only the placements. Rotation
// matters: Battlesnake's fixed start positions are not equivalent, so an
// unrotated result measures seat luck.
func Evaluate(b Batch, subject, opponent func(worker int) Searcher) Result {
	return run(b, func(worker int) table {
		return table{subject: subject(worker), opponent: opponent(worker), rotate: true}
	})
}

type table struct {
	subject, opponent Searcher
	record            bool
	rotate            bool
}

func run(b Batch, setup func(worker int) table) Result {
	games := make(chan int, b.Games)
	for g := 0; g < b.Games; g++ {
		games <- g
	}
	close(games)

	var (
		mu     sync.Mutex
		result Result
	)

	var wg sync.WaitGroup
	for w := 0; w < b.Workers; w++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			rnd := rand.New(rand.NewSource(b.Seed + int64(worker)*7919))
			tbl := setup(worker)

			for g := range games {
				seat := 0
				if tbl.rotate {
					seat = g % b.Seats
				}
				samples, placement, frames := playGame(b, tbl, seat, rnd, g == 0)

				mu.Lock()
				result.Samples = append(result.Samples, samples...)
				result.Placement += placement[seat]
				result.Ranks[rankOf(placement, seat, b.Seats)]++
				result.Games++
				if frames != nil {
					result.Frames = frames
				}
				mu.Unlock()

				if b.Progress != nil {
					mu.Lock()
					done := result.Games
					mu.Unlock()
					b.Progress(done)
				}
			}
		}(w)
	}
	wg.Wait()

	return result
}

// playGame plays one game out. subjectSeat is the seat the subject occupies;
// every other seat is the opponent's.
func playGame(b Batch, tbl table, subjectSeat int, rnd *rand.Rand, keepFrames bool) ([]constrictor.Sample, [constrictor.MaxSnakes]float64, []constrictor.State) {
	s := constrictor.New(b.Board, b.Board, constrictor.Starts(b.Board, b.Board, b.Seats, rnd))

	var (
		samples []constrictor.Sample
		seats   []int // sample i belongs to seats[i], so placements can be stamped at the end
		frames  []constrictor.State
	)
	if keepFrames {
		frames = []constrictor.State{*s}
	}

	for !s.Over() {
		moves, policy := tbl.subject.Search(s)

		if tbl.subject != tbl.opponent {
			// Two different bots, so two searches: each side plans with its own
			// network, and neither sees the other's tree.
			opponentMoves, _ := tbl.opponent.Search(s)
			for seat := 0; seat < b.Seats; seat++ {
				if seat != subjectSeat {
					moves[seat] = opponentMoves[seat]
				}
			}
		}

		if tbl.record {
			for seat := 0; seat < b.Seats; seat++ {
				if s.IsAlive(seat) {
					samples = append(samples, constrictor.NewSample(s, seat, policy[seat]))
					seats = append(seats, seat)
				}
			}
		}

		s.Apply(moves)
		if keepFrames {
			frames = append(frames, *s)
		}
	}

	placement := s.Placement()
	for i := range samples {
		samples[i].Value = float32(placement[seats[i]])
	}
	return samples, placement, frames
}

// rankOf returns the seat's finishing position, 0 for first.
func rankOf(placement [constrictor.MaxSnakes]float64, seat, seats int) int {
	better := 0
	for i := 0; i < seats; i++ {
		if placement[i] > placement[seat] {
			better++
		}
	}
	return better
}
