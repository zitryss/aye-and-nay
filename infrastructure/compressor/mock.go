package compressor

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"os"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/pool"
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
		case multipart.File:
			_ = v.Close()
		case *bytes.Buffer:
			pool.PutBuffer(v)
		default:
			panic(errors.Wrap(domain.ErrUnknown))
		}
	}()
	buf := pool.GetBuffer()
	n, err := io.Copy(buf, f)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	return model.File{Reader: buf, Size: n}, nil
}
