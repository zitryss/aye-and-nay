package compressor

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/internal/pool"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func NewMock() *Mock {
	return &Mock{}
}

type Mock struct {
}

func (m *Mock) Compress(_ context.Context, f model.File) (model.File, error) {
	defer func() {
		switch v := f.Reader.(type) {
		case *os.File:
			_ = v.Close()
			_ = os.Remove(v.Name())
		case *bytes.Buffer:
			pool.PutBuffer(v)
		}
	}()
	buf := pool.GetBuffer()
	n, err := io.CopyN(buf, f, f.Size)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	return model.File{Reader: buf, Size: n}, nil
}

func NewFail(opts ...options) *fail {
	sp := NewShortPixel(opts...)
	return &fail{sp}
}

type fail struct {
	*Shortpixel
}

func (f *fail) compress(_ context.Context, _ model.File) (model.File, error) {
	return model.File{}, errors.Wrap(model.ErrThirdPartyUnavailable)
}
