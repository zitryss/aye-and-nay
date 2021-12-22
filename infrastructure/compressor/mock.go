package compressor

import (
	"context"
	"io"

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
	defer f.Close()
	buf := pool.GetBufferN(f.Size)
	n, err := io.Copy(buf, f.Reader)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	closeFn := func() error {
		pool.PutBuffer(buf)
		return nil
	}
	return model.NewFile(buf, closeFn, n), nil
}
