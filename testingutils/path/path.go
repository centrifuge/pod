//go:build unit || integration || testworld

package path

import (
	"path/filepath"
	"runtime"
)

var (
	_, currentFilePath, _, _ = runtime.Caller(0)

	ProjectRoot = filepath.Join(filepath.Dir(currentFilePath), "../..")
)

func AppendPathToProjectRoot(path string) string {
	return filepath.Join(ProjectRoot, path)
}
