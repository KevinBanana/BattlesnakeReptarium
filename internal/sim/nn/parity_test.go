package nn

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"BattlesnakeReptarium/internal/sim/constrictor"
)

// Stage 3 of the Phase 1 gate. Stages 1 and 2 - the Go and Python encoders
// agreeing, and PyTorch and ONNX Runtime agreeing - run in training/parity.py.
// This is the third: the path that actually serves moves produces what PyTorch
// produced, on the same positions.
//
//	go test ./internal/sim/constrictor -run WriteParityFixture
//	.venv/Scripts/python training/parity.py
//	go test ./internal/sim/nn -run Parity
const (
	tolerance    = 1e-4
	fixturePath  = "../constrictor/testdata/encode_parity.json"
	modelPath    = "../../../training/testdata/parity.onnx"
	expectedPath = "../../../training/testdata/parity_expected.json"
	libraryGlob  = "../../../third_party/onnxruntime-*/lib/onnxruntime.dll"
)

type fixture struct {
	Samples []struct {
		Width  int   `json:"w"`
		Height int   `json:"h"`
		Seats  int   `json:"n"`
		Turn   int   `json:"turn"`
		Alive  uint8 `json:"alive"`
		Cells  []int `json:"cells"`
		Heads  []int `json:"heads"`
		Ego    int   `json:"ego"`
	} `json:"samples"`
}

type expected struct {
	Policy [][]float32 `json:"policy"`
	Value  []float32   `json:"value"`
}

func TestParityWithPyTorch(t *testing.T) {
	fix := loadFixture(t)
	want := loadExpected(t)
	require.Len(t, want.Policy, len(fix.Samples), "expected outputs do not match the fixture - regenerate both")

	// Encoding happens here rather than being read back from the fixture: the
	// point is to run the same code the search will run, not to re-check a
	// number Go already wrote.
	board := fix.Samples[0].Width
	stride := constrictor.EncodeLen(board, board)
	in := make([]float32, 0, len(fix.Samples)*stride)
	for _, s := range fix.Samples {
		buf := make([]float32, stride)
		state(s.Cells, s.Heads, s.Alive, s.Turn, s.Width, s.Height, s.Seats).Encode(s.Ego, buf)
		in = append(in, buf...)
	}

	session := open(t, board)
	defer session.Close()

	policy, value, err := session.Run(in, len(fix.Samples))
	require.NoError(t, err)

	worstPolicy, worstValue := 0.0, 0.0
	for i := range fix.Samples {
		for m := 0; m < MoveCount; m++ {
			worstPolicy = max(worstPolicy, absDiff(policy[i*MoveCount+m], want.Policy[i][m]))
		}
		worstValue = max(worstValue, absDiff(value[i], want.Value[i]))
	}

	t.Logf("%d positions: worst policy %.2e, worst value %.2e", len(fix.Samples), worstPolicy, worstValue)
	require.Less(t, worstPolicy, tolerance, "policy differs from PyTorch")
	require.Less(t, worstValue, tolerance, "value differs from PyTorch")
}

// state rebuilds a position from the fixture's fields.
func state(cells, heads []int, alive uint8, turn, w, h, seats int) *constrictor.State {
	s := &constrictor.State{W: w, H: h, N: seats, Turn: turn, Alive: alive}
	for i := range s.Cells {
		s.Cells[i] = constrictor.Empty
	}
	for c, owner := range cells {
		s.Cells[c] = uint8(owner)
	}
	for i, head := range heads {
		s.Heads[i] = uint8(head)
	}
	return s
}

func open(t *testing.T, board int) *Session {
	t.Helper()

	matches, err := filepath.Glob(libraryGlob)
	require.NoError(t, err)
	if len(matches) == 0 {
		t.Skipf("no ONNX Runtime library matching %s - see training/requirements.txt", libraryGlob)
	}
	t.Setenv(LibraryEnv, matches[0])
	
	session, err := Open(modelPath, constrictor.Planes, board, board)
	require.NoError(t, err)
	return session
}

func loadFixture(t *testing.T) fixture {
	t.Helper()
	var f fixture
	readJSON(t, fixturePath, &f, "go test ./internal/sim/constrictor -run WriteParityFixture")
	require.NotEmpty(t, f.Samples)
	return f
}

func loadExpected(t *testing.T) expected {
	t.Helper()
	var e expected
	readJSON(t, expectedPath, &e, "python training/parity.py")
	return e
}

func readJSON(t *testing.T, path string, into any, regenerate string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skipf("%s missing - run: %s", path, regenerate)
	}
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, into))
}

func absDiff(a, b float32) float64 {
	d := float64(a) - float64(b)
	if d < 0 {
		return -d
	}
	return d
}
