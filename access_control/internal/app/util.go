package app

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
)

func validateFileMimeType(file multipart.File) error {
	bytes, err := io.ReadAll(file) // :(((
	if err != nil {
		return &ImageAppError{
			Code:    Internal,
			Message: fmt.Sprintf("Internal error. cannot read file data into memory %v", err),
		}
	}
	mimeType := http.DetectContentType(bytes)

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
