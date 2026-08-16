package nn

import (
	"path/filepath"
	"testing"

	"BattlesnakeReptarium/internal/sim/constrictor"
)

// openForBench is open() for benchmarks, which have no *testing.T.
func openForBench(b *testing.B, board int) *Session {
	b.Helper()

	matches, err := filepath.Glob(libraryGlob)
	if err != nil || len(matches) == 0 {
		b.Skipf("no ONNX Runtime library matching %s", libraryGlob)
	}
	b.Setenv(LibraryEnv, matches[0])

	session, err := Open(modelPath, constrictor.Planes, board, board)
	if err != nil {
		b.Skipf("no model at %s: run training/parity.py (%v)", modelPath, err)
	}
	return session
}
