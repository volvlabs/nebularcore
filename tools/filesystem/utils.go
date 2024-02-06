package filesystem

import (
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func GetRootDir(relativePath string) string {
	_, currentFile, _, _ := runtime.Caller(1)
	return filepath.Join(path.Dir(currentFile), relativePath)
}

func IsValidAudioFileType(fileHeader *multipart.FileHeader) bool {
	file, err := fileHeader.Open()
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return false
	}

	file.Seek(0, 0)

	mimeType := http.DetectContentType(buffer)

	return strings.HasPrefix(mimeType, "audio/")
}
