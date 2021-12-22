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

func NewImaginary() (*Imaginary, error) {
	conf := newImaginaryConfig()
	ctx, cancel := context.WithTimeout(context.Background(), conf.timeout)
	defer cancel()
	err := retry.Do(conf.times, conf.pause, func() error {
		url := "http://" + conf.host + ":" + conf.port + "/health"
		body := io.Reader(nil)
		req, err := http.NewRequestWithContext(ctx, "GET", url, body)
		if err != nil {
			return errors.Wrap(err)
		}
		c := http.Client{}
		resp, err := c.Do(req)
		if err != nil {
			return errors.Wrap(err)
		}
		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return errors.Wrap(err)
		}
		err = resp.Body.Close()
		if err != nil {
			return errors.Wrap(err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return errors.Wrap(errors.New("no connection to imaginary"))
		}
		return nil
	})
	if err != nil {
		return &Imaginary{}, errors.Wrap(err)
	}
	return &Imaginary{conf}, nil
}

type Imaginary struct {
	conf imaginaryConfig
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
	url := "http://" + im.conf.host + ":" + im.conf.port + "/convert?type=png&compression=9"
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	req.Header.Set("Content-Type", multi.FormDataContentType())
	c := http.Client{Timeout: im.conf.timeout}
	resp := (*http.Response)(nil)
	err = retry.Do(im.conf.times, im.conf.pause, func() error {
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
