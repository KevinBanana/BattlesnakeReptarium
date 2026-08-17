// Package nn evaluates the trained network.
//
// It sits under internal/sim beside the gamemode packages because it belongs to
// the search, not to the service: nothing in model, server or controllers knows
// it exists, and nothing here knows about HTTP.
//
// It is also the only package that touches ONNX Runtime, and therefore the only
// one that needs cgo. The gamemode packages next door stay pure Go and keep
// building on a machine with no C toolchain at all.
//
// A trained 4-block ResNet is a few hundred thousand parameters of convolution.
// Hand-rolling that forward pass in Go would be the code nobody wants to debug
// at 3am, and it would be a second implementation of the network to keep in
// step with PyTorch's. One inference path, no train/serve skew.
package nn

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

// LibraryEnv names the ONNX Runtime shared library to load. Unset, the copy in
// third_party is used - see locate.go for why that beats the system search
// path.
const LibraryEnv = "ONNXRUNTIME_LIB"

// MoveCount is the policy head's width: one logit per direction.
const MoveCount = 4

var (
	initOnce sync.Once
	initErr  error
)

// initialize brings up the ORT environment once per process. The library has to
// be chosen before the environment exists, so both live here, and the device is
// fixed for the process - a mixed CPU and GPU run would need two libraries
// loaded at once, which ORT does not do.
func initialize(cuda bool) {
	if cuda {
		addCUDALibraries()
	}

	path := os.Getenv(LibraryEnv)
	if path == "" {
		path = findLibrary(cuda)
	}
	if path != "" {
		ort.SetSharedLibraryPath(path)
	}

	initErr = ort.InitializeEnvironment()
}

// Session is a loaded network, safe to reuse across evaluations.
type Session struct {
	session      *ort.DynamicAdvancedSession
	planes, w, h int
}

// Options select the device. The zero value is CPU, which is what serving
// wants: a single game needs one small evaluation back quickly, and a GPU is no
// faster at that.
type Options struct {
	// CUDA runs on the GPU. Only worth it behind a Server, where batches are
	// large enough to pay off the fixed cost of a call, and it needs the GPU
	// build of the ONNX Runtime library rather than the CPU one.
	CUDA     bool
	DeviceID int

	// Threads is ONNX Runtime's intra-op pool: how many threads may work on a
	// single evaluation. Zero means one, which is right whenever the caller
	// provides the parallelism - self-play runs a game per core, and letting ORT
	// size its own pool on top of that oversubscribes badly. Serving is the
	// case where it might pay: one move is one goroutine, so the other cores
	// are idle anyway.
	Threads int
}

// Open loads a model exported by training/net.py, shaped for planes planes of
// w x h - constrictor.Planes and the board it was trained on.
func Open(modelPath string, planes, w, h int) (*Session, error) {
	return OpenWith(modelPath, planes, w, h, Options{})
}

// OpenWith is Open with a device choice.
func OpenWith(modelPath string, planes, w, h int, opts Options) (*Session, error) {
	initOnce.Do(func() { initialize(opts.CUDA) })
	if initErr != nil {
		return nil, fmt.Errorf("onnxruntime: %w (no library found under third_party; set %s)", initErr, LibraryEnv)
	}

	// One thread per Run unless asked otherwise - see Options.Threads. Measured
	// at 16 self-play workers, one thread changes throughput by nothing at all;
	// it is kept because it is the correct setting for the way self-play is
	// parallelised, not because it bought anything.
	options, err := ort.NewSessionOptions()
	if err != nil {
		return nil, err
	}
	defer options.Destroy()
	threads := opts.Threads
	if threads <= 0 {
		threads = 1
	}
	if err := options.SetIntraOpNumThreads(threads); err != nil {
		return nil, err
	}
	if err := options.SetInterOpNumThreads(1); err != nil {
		return nil, err
	}

	if opts.CUDA {
		cuda, err := ort.NewCUDAProviderOptions()
		if err != nil {
			return nil, fmt.Errorf("cuda provider (is this the GPU build of the library?): %w", err)
		}
		defer cuda.Destroy()
		if err := cuda.Update(map[string]string{"device_id": strconv.Itoa(opts.DeviceID)}); err != nil {
			return nil, err
		}
		if err := options.AppendExecutionProviderCUDA(cuda); err != nil {
			return nil, fmt.Errorf("enabling cuda: %w", err)
		}
	}

	session, err := ort.NewDynamicAdvancedSession(modelPath,
		[]string{"board"}, []string{"policy", "value"}, options)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", modelPath, err)
	}
	return &Session{session: session, planes: planes, w: w, h: h}, nil
}

func (s *Session) Close() error { return s.session.Destroy() }

// Run evaluates batch positions at once. in holds them back to back, each
// laid out by State.Encode. policy is batch*MoveCount logits and value is one
// placement estimate per position.
//
// Batching is the point: a search node is evaluated once per seat, so the
// natural call is a batch of 4, which costs about what a batch of 1 does.
//
// Allocating fresh tensors per call looks wasteful and is not: measured across
// batch sizes 1, 4, 16 and 64, cost is flat at ~1.2ms per position, so there is
// no per-call overhead to hoist. The time is arithmetic - a 4x64 trunk over 121
// cells - and the only ways down are a smaller network, fewer simulations, or a
// device that does this much faster.
func (s *Session) Run(in []float32, batch int) (policy []float32, value []float32, err error) {
	if want := batch * s.planes * s.w * s.h; len(in) != want {
		return nil, nil, fmt.Errorf("expected %d floats for a batch of %d, got %d", want, batch, len(in))
	}

	input, err := ort.NewTensor(ort.NewShape(int64(batch), int64(s.planes), int64(s.h), int64(s.w)), in)
	if err != nil {
		return nil, nil, err
	}
	defer input.Destroy()

	policyTensor, err := ort.NewEmptyTensor[float32](ort.NewShape(int64(batch), MoveCount))
	if err != nil {
		return nil, nil, err
	}
	defer policyTensor.Destroy()

	valueTensor, err := ort.NewEmptyTensor[float32](ort.NewShape(int64(batch)))
	if err != nil {
		return nil, nil, err
	}
	defer valueTensor.Destroy()

	if err := s.session.Run([]ort.Value{input}, []ort.Value{policyTensor, valueTensor}); err != nil {
		return nil, nil, err
	}

	// The tensors own C memory that Destroy frees, so hand back copies.
	policy = append([]float32(nil), policyTensor.GetData()...)
	value = append([]float32(nil), valueTensor.GetData()...)
	return policy, value, nil
}
