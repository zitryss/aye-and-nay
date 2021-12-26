package storage

import (
	"context"
	"io"

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
	defer f.Close()
	albumB64 := base64.FromUint64(album)
	imageB64 := base64.FromUint64(image)
	filename := "albums/" + albumB64 + "/images/" + imageB64
	_, _ = io.Copy(io.Discard, f.Reader)
	src := "/aye-and-nay/" + filename
	return src, nil
}

func (m *Mock) Get(_ context.Context, _ uint64, _ uint64) (model.File, error) {
	f := Png()
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

func (m *Mock) Remove(_ context.Context, _ uint64, _ uint64) error {
	return nil
}

func (m *Mock) Health(_ context.Context) (bool, error) {
	return true, nil
}
