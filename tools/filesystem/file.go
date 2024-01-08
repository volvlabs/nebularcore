package filesystem

import (
	"path"
	"path/filepath"
	"runtime"
)

func GetRootDir(relativePath string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(path.Dir(currentFile), relativePath)
}
