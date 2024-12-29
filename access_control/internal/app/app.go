package app

import (
	"io"
	"os"
)

type App interface {
	LoadImage(path string) (*os.File, error)
	SaveImage(w io.Writer, path string) error
}
