package model

import (
	"io"
)

type File struct {
	io.Reader
	Size int64
}
