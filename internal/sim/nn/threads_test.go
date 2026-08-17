package nn

import (
	"fmt"
	"path/filepath"
	"testing"

	"BattlesnakeReptarium/internal/sim/constrictor"
)

// BenchmarkThreads asks whether letting ONNX Runtime split one evaluation
// across cores is worth anything at serving sizes.
//
// Serving is where it could pay: a move is one goroutine descending one tree,
// so the other fifteen threads are idle while the deadline runs out. Batch 4 is
// what a four-snake node costs; batch 2 is a duel, which is most of the late
// game.
func BenchmarkThreads(b *testing.B) {
	const board = 11

	matches, err := filepath.Glob(libraryGlob)
	if err != nil || len(matches) == 0 {
		b.Skipf("no ONNX Runtime library matching %s", libraryGlob)
	}

	for _, batch := range []int{2, 4} {
		for _, threads := range []int{1, 2, 4, 8} {
			b.Run(fmt.Sprintf("batch-%d/threads-%d", batch, threads), func(b *testing.B) {
				b.Setenv(LibraryEnv, matches[0])

				session, err := OpenWith(modelPath, constrictor.Planes, board, board, Options{Threads: threads})
				if err != nil {
					b.Skipf("no model at %s: run training/parity.py (%v)", modelPath, err)
				}
				defer session.Close()

				in := make([]float32, batch*constrictor.EncodeLen(board, board))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if _, _, err := session.Run(in, batch); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
