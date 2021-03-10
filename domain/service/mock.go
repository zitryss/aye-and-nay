package service

import (
	"context"
	"time"

	"github.com/zitryss/aye-and-nay/domain/model"
)

func NewMock(err error) *Mock {
	return &Mock{err}
}

type Mock struct {
	err error
}

func (m *Mock) Album(ctx context.Context, ff []model.File, dur time.Duration) (uint64, error) {
	if m.err != nil {
		return "", m.err
	}
	return "N2fxX5zbDh8RJQvx1", nil
}

func (m *Mock) Progress(ctx context.Context, album uint64) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
}

func (m *Mock) Pair(ctx context.Context, album uint64) (model.Image, model.Image, error) {
	if m.err != nil {
		return model.Image{}, model.Image{}, m.err
	}
	img1 := model.Image{Src: "/aye-and-nay/albums/DfsXRkDxVeH2xmme1/images/DfsXRkDxVeH2xmme2", Token: "DfsXRkDxVeH2xmme5"}
	img2 := model.Image{Src: "/aye-and-nay/albums/DfsXRkDxVeH2xmme1/images/DfsXRkDxVeH2xmme3", Token: "DfsXRkDxVeH2xmme6"}
	return img1, img2, nil
}

func (m *Mock) Vote(ctx context.Context, album uint64, tokenFrom uint64, tokenTo uint64) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *Mock) Top(ctx context.Context, album uint64) ([]model.Image, error) {
	if m.err != nil {
		return nil, m.err
	}
	img1 := model.Image{Src: "/aye-and-nay/albums/bYCppY8q6qjvXjMZ1/images/bYCppY8q6qjvXjMZ2", Rating: 0.5}
	img2 := model.Image{Src: "/aye-and-nay/albums/bYCppY8q6qjvXjMZ1/images/bYCppY8q6qjvXjMZ3", Rating: 0.5}
	imgs := []model.Image{img1, img2}
	return imgs, nil
}
