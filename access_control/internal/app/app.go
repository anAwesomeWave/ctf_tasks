package app

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
)

type ImageType int

const (
	Avatar       ImageType = 0
	DefaultImage ImageType = 1
)

type App interface {
	LoadImage(id string, iType ImageType) (*os.File, error)
	SaveImage(file multipart.File, iType ImageType) error
}

type DefaultApp struct {
	//basePathImages
	//basePathAvatars
}

func validateFile(fileBytes []byte) error {
	mimeType := http.DetectContentType(fileBytes)

	// Разрешённые типы файлов
	allowedMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	if !allowedMimeTypes[mimeType] {
		return fmt.Errorf("unsupported MIME type: %s", mimeType)
	}

	return nil
}
