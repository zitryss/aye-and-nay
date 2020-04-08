package compressor

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/model"
)

func NewMock() mock {
	return mock{}
}

type mock struct {
}

func (m *mock) Compress(_ context.Context, _ []model.Image) error {
	return nil
}

func NewFail() fail {
	return fail{}
}

type fail struct {
}

func (f *fail) Compress(_ context.Context, _ []model.Image) error {
	return model.ErrThirdPartyUnavailable
}
