package app

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
)

func validateFileMimeType(fileBytes []byte) error {
	mimeType := http.DetectContentType(fileBytes)

	allowedMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	if !allowedMimeTypes[mimeType] {
		return &ImageAppError{
			Code:    Image,
			Message: "IMAGE ERROR: Wrong image mime types",
		}
	}

	return nil
}

func isExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}
