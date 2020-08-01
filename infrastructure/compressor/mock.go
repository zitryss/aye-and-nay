package compressor

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func NewMock() mock {
	return mock{}
}

type mock struct {
}

func (m *mock) Compress(_ context.Context, b []byte) ([]byte, error) {
	return b, nil
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

func (f *fail) compress(_ context.Context, _ []byte) ([]byte, error) {
	return nil, errors.Wrap(model.ErrThirdPartyUnavailable)
}
