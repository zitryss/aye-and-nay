package compressor

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/pool"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

func NewImaginary(ctx context.Context, conf ImaginaryConfig) (*Imaginary, error) {
	im := &Imaginary{conf}
	ctx, cancel := context.WithTimeout(ctx, conf.Timeout)
	defer cancel()
	err := retry.Do(conf.RetryTimes, conf.RetryPause, func() error {
		_, err := im.Health(ctx)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return &Imaginary{}, errors.Wrap(err)
	}
	return im, nil
}

type Imaginary struct {
	conf ImaginaryConfig
}

func (im *Imaginary) Compress(ctx context.Context, f model.File) (model.File, error) {
	defer f.Close()
	buf := pool.GetBufferN(f.Size)
	tee := model.File{
		Reader: io.TeeReader(f.Reader, buf),
		Size:   f.Size,
	}
	body := pool.GetBufferN(f.Size)
	defer pool.PutBuffer(body)
	multi := multipart.NewWriter(body)
	part, err := multi.CreateFormFile("file", "non-empty-field")
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	n, err := io.Copy(part, tee.Reader)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	err = multi.Close()
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	url := "http://" + im.conf.Host + ":" + im.conf.Port + "/convert?type=png&compression=9"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	req.Header.Set("Content-Type", multi.FormDataContentType())
	c := http.Client{Timeout: im.conf.Timeout}
	resp := (*http.Response)(nil)
	err = retry.Do(im.conf.RetryTimes, im.conf.RetryPause, func() error {
		resp, err = c.Do(req)
		if err != nil {
			return errors.Wrapf(domain.ErrThirdPartyUnavailable, "%s", err)
		}
		if resp.StatusCode == 406 {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			return errors.Wrap(domain.ErrUnsupportedMediaType)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			return errors.Wrapf(domain.ErrThirdPartyUnavailable, "status code %d", resp.StatusCode)
		}
		return nil
	})
	if errors.Is(err, domain.ErrUnsupportedMediaType) {
		return model.File{Reader: buf, Size: n}, nil
	}
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	buf.Reset()
	n, err = io.Copy(buf, resp.Body)
	if err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return model.File{}, errors.Wrap(err)
	}
	err = resp.Body.Close()
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	closeFn := func() error {
		pool.PutBuffer(buf)
		return nil
	}
	return model.NewFile(buf, closeFn, n), nil
}

func (im *Imaginary) Health(ctx context.Context) (bool, error) {
	url := "http://" + im.conf.Host + ":" + im.conf.Port + "/health"
	body := io.Reader(nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, body)
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthCompressor, "%s", err)
	}
	c := http.Client{Timeout: im.conf.Timeout}
	resp, err := c.Do(req)
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthCompressor, "%s", err)
	}
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return false, errors.Wrapf(domain.ErrBadHealthCompressor, "%s", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthCompressor, "%s", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return false, errors.Wrapf(domain.ErrBadHealthCompressor, "%s", "no connection to imaginary")
	}
	return true, nil
}
