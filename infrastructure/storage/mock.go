package storage

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/model"
)

func NewMock() mock {
	return mock{}
}

type mock struct {
}

func (m *mock) Upload(_ context.Context, album string, imgs []model.Image) error {
	for i := range imgs {
		img := &imgs[i]
		img.Src = "/aye-and-nay/albums/" + album + "/images/" + img.Id
	}
	return nil
}
