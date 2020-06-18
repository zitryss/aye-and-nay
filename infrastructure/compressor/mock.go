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
	return fail{}
}

type fail struct {
	err error
}

func (f *fail) Compress(ctx context.Context, b []byte) ([]byte, error) {
	if errors.Is(f.err, model.ErrThirdPartyUnavailable) {
		return b, nil
	}
	bb := []byte(nil)
	bb, f.err = f.compress(ctx, b)
	return bb, errors.Wrap(f.err)
}

func (f *fail) compress(_ context.Context, _ []byte) ([]byte, error) {
	return nil, model.ErrThirdPartyUnavailable
}
