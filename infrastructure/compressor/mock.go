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

func NewMock() mock {
	return mock{}
}

type mock struct {
}

func (m *mock) Compress(_ context.Context, f model.File) (model.File, error) {
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

func NewFail() fail {
	conf := newShortPixelConfig()
	return fail{
		shortpixel: shortpixel{
			conf: conf,
			ch:   make(chan struct{}, 1),
		},
	}
}

type fail struct {
	shortpixel
}

func (f *fail) compress(_ context.Context, _ model.File) (model.File, error) {
	return model.File{}, errors.Wrap(model.ErrThirdPartyUnavailable)
}
