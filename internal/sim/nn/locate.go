package nn

// locate.go finds the native libraries so nothing has to be exported by hand.
//
// Two separate problems, both of which bite as "the specified procedure could
// not be found":
//
//   - Which onnxruntime library. With none named, the loader takes whatever
//     onnxruntime.dll it finds on the system path, which on Windows is often an
//     unrelated one that has no OrtGetApiBase at all.
//   - Where CUDA's libraries are. The GPU build's provider needs cuBLAS, cuDNN
//     and the CUDA runtime, and the copies torch ships are a different version
//     than the ones ONNX Runtime was built against.
//
// Both are answered by things already in the repo, so answering them here beats
// documenting an export nobody remembers.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// libraryNames is what the shared library is called per platform.
var libraryNames = map[string]string{
	"windows": "onnxruntime.dll",
	"linux":   "libonnxruntime.so",
	"darwin":  "libonnxruntime.dylib",
}

// findLibrary returns the bundled ONNX Runtime matching the device, or "" if
// there is none to find. The GPU build is a separate download and lives in its
// own directory, distinguished by "gpu" in the name.
func findLibrary(cuda bool) string {
	root := moduleRoot()
	if root == "" {
		return ""
	}

	name := libraryNames[runtime.GOOS]
	if name == "" {
		return ""
	}

	matches, err := filepath.Glob(filepath.Join(root, "third_party", "onnxruntime-*", "lib", name))
	if err != nil {
		return ""
	}
	for _, match := range matches {
		if strings.Contains(filepath.Base(filepath.Dir(filepath.Dir(match))), "gpu") == cuda {
			return match
		}
	}
	return ""
}

// addCUDALibraries puts the venv's nvidia libraries where the provider will
// look. Setting PATH inside the process is enough because the provider library
// is loaded later, when the execution provider is appended.
func addCUDALibraries() {
	root := moduleRoot()
	if root == "" {
		return
	}

	var dirs []string
	for _, pattern := range [][]string{
		{".venv", "Lib", "site-packages", "nvidia", "*", "bin"},            // windows
		{".venv", "lib", "python*", "site-packages", "nvidia", "*", "lib"}, // posix
	} {
		found, err := filepath.Glob(filepath.Join(append([]string{root}, pattern...)...))
		if err == nil {
			dirs = append(dirs, found...)
		}
	}
	if len(dirs) == 0 {
		return
	}

	key := "PATH"
	if runtime.GOOS != "windows" {
		key = "LD_LIBRARY_PATH"
	}
	os.Setenv(key, strings.Join(dirs, string(os.PathListSeparator))+string(os.PathListSeparator)+os.Getenv(key))
}

// moduleRoot walks up from the working directory to the directory holding
// go.mod, so this works from the repo root and from a package directory during
// go test alike.
func moduleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
