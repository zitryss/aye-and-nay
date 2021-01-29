package storage

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/internal/pool"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func NewMock() *Mock {
	return &Mock{}
}

type Mock struct {
}

func (m *Mock) Put(_ context.Context, album string, image string, f model.File) (string, error) {
	defer func() {
		switch v := f.Reader.(type) {
		case *os.File:
			_ = v.Close()
			_ = os.Remove(v.Name())
		case *bytes.Buffer:
			pool.PutBuffer(v)
		}
	}()
	filename := "albums/" + album + "/images/" + image
	src := "/aye-and-nay/" + filename
	return src, nil
}

func (m *Mock) Get(_ context.Context, _ string, _ string) (model.File, error) {
	buf := pool.GetBuffer()
	f := Png()
	n, err := io.CopyN(buf, f, f.Size)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	return model.File{Reader: buf, Size: n}, nil
}

func (m *Mock) Remove(_ context.Context, _ string, _ string) error {
	return nil
}
