package compressor

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/model"
)

func NewImaginary() (*Imaginary, error) {
	return &Imaginary{}, nil
}

type Imaginary struct {
}

func (im *Imaginary) ping() error {
	return nil
}

func (im *Imaginary) Compress(ctx context.Context, f model.File) (model.File, error) {
	return model.File{}, nil
}

func (im *Imaginary) upload(ctx context.Context, f model.File) (string, error) {
	return "", nil
}

func (im *Imaginary) download(ctx context.Context, src string) (model.File, error) {
	return model.File{}, nil
}
