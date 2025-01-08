package app

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
)

type ImageType int

const (
	Avatar       ImageType = 0
	DefaultImage ImageType = 1
)

type ErrorCode int

const (
	Path     ErrorCode = 1
	Image    ErrorCode = 2
	Internal ErrorCode = 3
)

type ImageAppError struct {
	Code    ErrorCode
	Message string
}

func (e *ImageAppError) Error() string {
	return e.Message
}

type App interface {
	LoadImage(id string, index uint64, iType ImageType) (*os.File, error)
	SaveImage(file multipart.File, id string, index uint64, iType ImageType) error
}

type DefaultApp struct {
	basePathImages  string
	basePathAvatars string
}

func isPathValid(path string) bool {
	return !(strings.Contains(path, "..") || strings.HasPrefix(path, "/"))
}

func NewDefaultApp(imagesDir, avatarsDir string) (*DefaultApp, error) {
	if !isPathValid(imagesDir) || !isPathValid(avatarsDir) {
		return nil, &ImageAppError{
			Code: Path,
			Message: fmt.Sprintf(
				"PATH ERROR: Wrong image/avatar folder. Check ImageDir=%s, avatarDir=%s",
				imagesDir,
				avatarsDir,
			),
		}
	}
	return &DefaultApp{
		imagesDir,
		avatarsDir,
	}, nil
}

func (d DefaultApp) SaveImage(file multipart.File, id string, index uint64, iType ImageType) error {
	uploadPrefix := ""
	switch iType {
	case Avatar:
		uploadPrefix = d.basePathAvatars
	case DefaultImage:
		uploadPrefix = d.basePathImages
	default:
		return &ImageAppError{
			Code: Internal,
			Message: fmt.Sprintf(
				"INTERNAL ERROR: Incorrect Image iType value = %d",
				iType,
			),
		}
	}
	// 1. check if user's dir exists. create if necessary
	// 2. check if file doesn't exist
	// 3. store file
	uploadPrefix += id + "/"
	if ex, err := isExists(uploadPrefix); err != nil || !ex {
		if err != nil {
			return &ImageAppError{
				Code: Path,
				Message: fmt.Sprintf(
					"PATH ERROR: upload prefix error %s: %s",
					uploadPrefix,
					err.Error(),
				),
			}
		}
		// non-existing dir
		if err := os.Mkdir(uploadPrefix, 0744); err != nil {
			return &ImageAppError{
				Code: Internal,
				Message: fmt.Sprintf(
					"INTERNAL ERROR: user id folder creating error: %s",
					err.Error(),
				),
			}
		}
	}
	uploadPrefix += strconv.FormatUint(index, 10) + ".jpeg"
	if ex, err := isExists(uploadPrefix); err != nil || ex {
		if err != nil {
			return &ImageAppError{
				Code: Path,
				Message: fmt.Sprintf(
					"PATH ERROR: Error checking for image %s to be existed or not: %s",
					uploadPrefix,
					err.Error(),
				),
			}
		}
		return &ImageAppError{
			Code: Path,
			Message: fmt.Sprintf(
				"PATH ERROR: image %s already exists",
				uploadPrefix,
			),
		}
	}
	dst, err := os.Create(uploadPrefix)
	if err != nil {
		return &ImageAppError{
			Code: Path,
			Message: fmt.Sprintf("PATH ERROR: Unable to save image with path %s: %s",
				uploadPrefix,
				err.Error(),
			),
		}
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return &ImageAppError{
			Code: Image,
			Message: fmt.Sprintf("IMAGE ERROR: Unable to copy image data to file %s: %s",
				dst,
				err.Error(),
			),
		}
	}
	return nil
}

func (d DefaultApp) loadDefaultImage(id string, index uint64) (*os.File, error) {
	imagePath := d.basePathImages + id + "/" + strconv.FormatUint(index, 10) + ".jpeg"
	if ex, err := isExists(imagePath); err != nil || !ex {
		if err != nil {
			return nil, &ImageAppError{
				Code:    Path,
				Message: err.Error(),
			}
		}
		return nil, &ImageAppError{
			Code:    Path,
			Message: fmt.Sprintf("PATH ERROR: file %s doesn't exist", imagePath),
		}
	}
	// check rights here (add params) isOwner, isAdmin, isPublic
	return os.Open(imagePath)
}

func (d DefaultApp) loadAvatar(id string, index uint64) (*os.File, error) {
	avatarsPath := d.basePathAvatars + id + "/" + strconv.FormatUint(index, 10) + ".jpeg"
	if ex, err := isExists(avatarsPath); err != nil || !ex {
		if err != nil {
			return nil, &ImageAppError{
				Code:    Path,
				Message: err.Error(),
			}
		}
		return nil, &ImageAppError{
			Code:    Path,
			Message: fmt.Sprintf("PATH ERROR: file %s doesn't exist", avatarsPath),
		}
	}
	// check rights here (add params) isOwner, isAdmin, isPublic
	return os.Open(avatarsPath)
}

func (d DefaultApp) LoadImage(id string, index uint64, iType ImageType) (*os.File, error) {
	switch iType {
	case DefaultImage:
		return d.loadDefaultImage(id, index)
	case Avatar:
		return d.loadAvatar(id, index)
	default:
		return nil, &ImageAppError{
			Code:    Internal,
			Message: fmt.Sprintf("Image type is invalid. %s", iType),
		}

	}
}
