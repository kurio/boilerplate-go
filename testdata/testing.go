package testdata

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
)

// basepath is the root directory of this package
var basepath string

func init() {
	_, currentFile, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(currentFile)
}

// path returns the absolute path the given relative file or directory path,
// relative to the github.com/KurioApp/avalon directory in the user's GOPATH.
// If rel is already absolute, it is returned unmodified.
func path(relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}

	return filepath.Join(basepath, relPath)
}

// GetGolden is a function to get golden file
func GetGolden(t *testing.T, filename string) []byte {
	t.Helper()

	b, err := ioutil.ReadFile(path(filename + ".golden"))
	if err != nil {
		t.Fatal(err)
	}

	return b
}
