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
	defer func() { _ = file.Close() }()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return false
	}

	if _, err = file.Seek(0, 0); err != nil {
		return false
	}

	mimeType := http.DetectContentType(buffer)

	return strings.HasPrefix(mimeType, "audio/")
}

func EncodeFilePathAsFileURL(path string) string {
	slashPath := filepath.ToSlash(path)
	fileURL := "file:///"
	if slashPath[1] == ':' {
		fileURL += slashPath
	} else {
		fileURL += slashPath[1:]
	}
	return fileURL
}
