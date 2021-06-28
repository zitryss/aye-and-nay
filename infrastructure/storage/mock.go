package storage

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"os"

	"github.com/zitryss/aye-and-nay/domain/model"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/base64"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/pool"
)

func NewMock() *Mock {
	return &Mock{}
}

type Mock struct {
}

func (m *Mock) Put(_ context.Context, album uint64, image uint64, f model.File) (string, error) {
	defer func() {
		switch v := f.Reader.(type) {
		case *os.File:
			_ = v.Close()
			_ = os.Remove(v.Name())
		case multipart.File:
			_ = v.Close()
		case *bytes.Buffer:
			pool.PutBuffer(v)
		default:
			panic(errors.Wrap(model.ErrUnknown))
		}
	}()
	albumB64 := base64.FromUint64(album)
	imageB64 := base64.FromUint64(image)
	filename := "albums/" + albumB64 + "/images/" + imageB64
	src := "/aye-and-nay/" + filename
	return src, nil
}

func (m *Mock) Get(_ context.Context, _ uint64, _ uint64) (model.File, error) {
	buf := pool.GetBuffer()
	f := Png()
	n, err := io.CopyN(buf, f, f.Size)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	return model.File{Reader: buf, Size: n}, nil
}

func (m *Mock) Remove(_ context.Context, _ uint64, _ uint64) error {
	return nil
}
