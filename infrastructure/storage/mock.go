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

func NewMock() mock {
	return mock{}
}

type mock struct {
}

func (m *mock) Put(_ context.Context, album string, image string, f model.File) (string, error) {
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

func (m *mock) Get(_ context.Context, _ string, _ string) (model.File, error) {
	buf := pool.GetBuffer()
	f := Png()
	n, err := io.CopyN(buf, f, f.Size)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	return model.File{Reader: buf, Size: n}, nil
}

func (m *mock) Remove(_ context.Context, _ string, _ string) error {
	return nil
}
