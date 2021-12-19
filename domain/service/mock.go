package service

import (
	"context"
	"io"
	"time"

	"github.com/zitryss/aye-and-nay/domain/model"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/pool"
)

func NewMock(err error) *Mock {
	return &Mock{err}
}

type Mock struct {
	err error
}

func (m *Mock) Album(_ context.Context, _ []model.File, _ time.Duration) (uint64, error) {
	if m.err != nil {
		return 0x0, m.err
	}
	return 0x1BAD, nil
}

func (m *Mock) Progress(_ context.Context, _ uint64) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
}

func (m *Mock) Pair(_ context.Context, _ uint64) (model.Image, model.Image, error) {
	if m.err != nil {
		return model.Image{}, model.Image{}, m.err
	}
	img1 := model.Image{Src: "/aye-and-nay/albums/nkUAAAAAAAA/images/21EAAAAAAAA", Token: 0xC77F}
	img2 := model.Image{Src: "/aye-and-nay/albums/nkUAAAAAAAA/images/K2IAAAAAAAA", Token: 0xA989}
	return img1, img2, nil
}

func (m *Mock) Image(_ context.Context, _ uint64) (model.File, error) {
	if m.err != nil {
		return model.File{}, m.err
	}
	buf := pool.GetBuffer()
	f := Png()
	n, err := io.Copy(buf, f)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	return model.File{Reader: buf, Size: n}, nil
}

func (m *Mock) Vote(_ context.Context, _ uint64, _ uint64, _ uint64) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *Mock) Top(_ context.Context, _ uint64) ([]model.Image, error) {
	if m.err != nil {
		return nil, m.err
	}
	img1 := model.Image{Src: "/aye-and-nay/albums/byYAAAAAAAA/images/yFwAAAAAAAA", Rating: 0.5}
	img2 := model.Image{Src: "/aye-and-nay/albums/byYAAAAAAAA/images/jVgAAAAAAAA", Rating: 0.5}
	imgs := []model.Image{img1, img2}
	return imgs, nil
}
