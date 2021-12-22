package model

import (
	"io"
)

func NewFile(reader io.Reader, close func() error, size int64) File {
	return File{reader, close, size}
}

type File struct {
	io.Reader
	close func() error
	Size  int64
}

func (f File) Close() error {
	if f.close == nil {
		return nil
	}
	return f.close()
}
