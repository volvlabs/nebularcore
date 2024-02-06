package filesystem

import (
	"path"
	"path/filepath"
	"runtime"
)

func GetRootDir(relativePath string) string {
	_, currentFile, _, _ := runtime.Caller(1)
	return filepath.Join(path.Dir(currentFile), relativePath)
}
