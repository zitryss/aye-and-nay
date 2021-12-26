package compressor

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/pool"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

func NewShortPixel(opts ...options) *Shortpixel {
	conf := newShortPixelConfig()
	sp := &Shortpixel{
		conf: conf,
		ch:   make(chan struct{}, 1),
	}
	for _, opt := range opts {
		opt(sp)
	}
	return sp
}

type options func(*Shortpixel)

func WithHeartbeatRestart(ch chan<- interface{}) options {
	return func(sp *Shortpixel) {
		sp.heartbeat.restart = ch
	}
}

type Shortpixel struct {
	conf      shortPixelConfig
	done      uint32
	ch        chan struct{}
	heartbeat struct {
		restart chan<- interface{}
	}
}

func (sp *Shortpixel) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, sp.conf.timeout)
	defer cancel()
	_, err := sp.upload(ctx, Png())
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (sp *Shortpixel) Monitor() {
	go func() {
		for {
			<-sp.ch
			if sp.heartbeat.restart != nil {
				sp.heartbeat.restart <- struct{}{}
			}
			time.Sleep(sp.conf.restartIn)
			atomic.StoreUint32(&sp.done, 0)
			if sp.heartbeat.restart != nil {
				sp.heartbeat.restart <- struct{}{}
			}
		}
	}()
}

func (sp *Shortpixel) Compress(ctx context.Context, f model.File) (model.File, error) {
	defer f.Close()
	if atomic.LoadUint32(&sp.done) != 0 {
		buf := pool.GetBufferN(f.Size)
		n, err := io.Copy(buf, f.Reader)
		if err != nil {
			return model.File{}, errors.Wrap(err)
		}
		return model.File{Reader: buf, Size: n}, nil
	}
	buf, err := sp.compress(ctx, f)
	if errors.Is(err, domain.ErrThirdPartyUnavailable) {
		if atomic.CompareAndSwapUint32(&sp.done, 0, 1) {
			sp.ch <- struct{}{}
		}
		return model.File{}, errors.Wrap(err)
	}
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	return buf, nil
}

func (sp *Shortpixel) compress(ctx context.Context, f model.File) (model.File, error) {
	src, err := sp.upload(ctx, f)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	bb, err := sp.download(ctx, src)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	return bb, nil
}

func (sp *Shortpixel) upload(ctx context.Context, f model.File) (string, error) {
	body := pool.GetBufferN(f.Size)
	defer pool.PutBuffer(body)
	multi := multipart.NewWriter(body)
	part, err := multi.CreateFormField("key")
	if err != nil {
		return "", errors.Wrap(err)
	}
	_, err = io.WriteString(part, sp.conf.apiKey)
	if err != nil {
		return "", errors.Wrap(err)
	}
	part, err = multi.CreateFormField("lossy")
	if err != nil {
		return "", errors.Wrap(err)
	}
	_, err = io.WriteString(part, "1")
	if err != nil {
		return "", errors.Wrap(err)
	}
	part, err = multi.CreateFormField("wait")
	if err != nil {
		return "", errors.Wrap(err)
	}
	_, err = io.WriteString(part, sp.conf.wait)
	if err != nil {
		return "", errors.Wrap(err)
	}
	part, err = multi.CreateFormField("convertto")
	if err != nil {
		return "", errors.Wrap(err)
	}
	_, err = io.WriteString(part, "png")
	if err != nil {
		return "", errors.Wrap(err)
	}
	part, err = multi.CreateFormField("file_paths")
	if err != nil {
		return "", errors.Wrap(err)
	}
	_, err = io.WriteString(part, `{"file": ""}`)
	if err != nil {
		return "", errors.Wrap(err)
	}
	part, err = multi.CreateFormFile("file", "non-empty-field")
	if err != nil {
		return "", errors.Wrap(err)
	}
	_, err = io.Copy(part, f.Reader)
	if err != nil {
		return "", errors.Wrap(err)
	}
	err = multi.Close()
	if err != nil {
		return "", errors.Wrap(err)
	}
	c := http.Client{Timeout: sp.conf.uploadTimeout}
	req, err := http.NewRequestWithContext(ctx, "POST", sp.conf.url, body)
	if err != nil {
		return "", errors.Wrap(err)
	}
	req.Header.Set("Content-Type", multi.FormDataContentType())
	resp := (*http.Response)(nil)
	err = retry.Do(sp.conf.times, sp.conf.pause, func() error {
		resp, err = c.Do(req)
		if err != nil {
			return errors.Wrapf(domain.ErrThirdPartyUnavailable, "%s", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			return errors.Wrapf(domain.ErrThirdPartyUnavailable, "status code %d", resp.StatusCode)
		}
		return nil
	})
	if err != nil {
		return "", errors.Wrap(err)
	}
	buf := pool.GetBufferN(resp.ContentLength)
	defer pool.PutBuffer(buf)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return "", errors.Wrap(err)
	}
	bb := buf.Bytes()
	if bb[0] == 91 && bb[len(bb)-1] == 93 {
		bb = bb[1 : len(bb)-1]
	}
	if bb[0] == 91 && (bb[len(bb)-2] == 93 && bb[len(bb)-1] == 10) {
		bb = bb[1 : len(bb)-2]
	}
	buf.Reset()
	buf.Write(bb)
	response := struct {
		Status struct {
			Code    interface{}
			Message string
		}
		OriginalUrl string
		LossyUrl    string
	}{}
	err = json.NewDecoder(buf).Decode(&response)
	if err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return "", errors.Wrapf(domain.ErrThirdPartyUnavailable, "%s", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return "", errors.Wrap(err)
	}
	src := ""
	switch response.Status.Code {
	case "1":
		src, err = sp.repeat(ctx, response.OriginalUrl)
		if err != nil {
			return "", errors.Wrap(err)
		}
	case "2":
		src = response.LossyUrl
	case -201.0, -202.0:
		return "", errors.Wrap(domain.ErrNotImage)
	default:
		return "", errors.Wrapf(domain.ErrThirdPartyUnavailable, "status code %v: message %q", response.Status.Code, response.Status.Message)
	}
	return src, nil
}

func (sp *Shortpixel) repeat(ctx context.Context, src string) (string, error) {
	time.Sleep(sp.conf.repeatIn)
	body := pool.GetBuffer()
	defer pool.PutBuffer(body)
	request := struct {
		Key       string   `json:"key"`
		Lossy     string   `json:"lossy"`
		Wait      string   `json:"wait"`
		Convertto string   `json:"convertto"`
		Urllist   []string `json:"urllist"`
	}{
		Key:       sp.conf.apiKey,
		Lossy:     "1",
		Wait:      sp.conf.wait,
		Convertto: "png",
		Urllist:   []string{src},
	}
	err := json.NewEncoder(body).Encode(request)
	if err != nil {
		return "", errors.Wrap(err)
	}
	c := http.Client{Timeout: sp.conf.uploadTimeout}
	req, err := http.NewRequestWithContext(ctx, "POST", sp.conf.url2, body)
	if err != nil {
		return "", errors.Wrap(err)
	}
	resp := (*http.Response)(nil)
	err = retry.Do(sp.conf.times, sp.conf.pause, func() error {
		resp, err = c.Do(req)
		if err != nil {
			return errors.Wrapf(domain.ErrThirdPartyUnavailable, "%s", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			return errors.Wrapf(domain.ErrThirdPartyUnavailable, "status code %d", resp.StatusCode)
		}
		return nil
	})
	if err != nil {
		return "", errors.Wrap(err)
	}
	buf := pool.GetBufferN(resp.ContentLength)
	defer pool.PutBuffer(buf)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return "", errors.Wrap(err)
	}
	b := buf.Bytes()
	if b[0] == 91 && b[len(b)-1] == 93 {
		b = b[1 : len(b)-1]
	}
	if b[0] == 91 && (b[len(b)-2] == 93 && b[len(b)-1] == 10) {
		b = b[1 : len(b)-2]
	}
	buf.Reset()
	buf.Write(b)
	response := struct {
		Status struct {
			Code    interface{}
			Message string
		}
		LossyUrl string
	}{}
	err = json.NewDecoder(buf).Decode(&response)
	if err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return "", errors.Wrapf(domain.ErrThirdPartyUnavailable, "%s", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return "", errors.Wrap(err)
	}
	switch response.Status.Code {
	case "1":
		return "", errors.Wrapf(domain.ErrThirdPartyUnavailable, "status code %v: message %q", response.Status.Code, response.Status.Message)
	case "2":
		return response.LossyUrl, nil
	default:
		return "", errors.Wrapf(domain.ErrThirdPartyUnavailable, "status code %v: message %q", response.Status.Code, response.Status.Message)
	}
}

func (sp *Shortpixel) download(ctx context.Context, src string) (model.File, error) {
	c := http.Client{Timeout: sp.conf.downloadTimeout}
	body := io.Reader(nil)
	req, err := http.NewRequestWithContext(ctx, "GET", src, body)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	resp := (*http.Response)(nil)
	err = retry.Do(sp.conf.times, sp.conf.pause, func() error {
		resp, err = c.Do(req)
		if err != nil {
			return errors.Wrapf(domain.ErrThirdPartyUnavailable, "%s", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			return errors.Wrapf(domain.ErrThirdPartyUnavailable, "status code %d", resp.StatusCode)
		}
		return nil
	})
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	buf := pool.GetBufferN(resp.ContentLength)
	n, err := io.Copy(buf, resp.Body)
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

func (sp *Shortpixel) Health(ctx context.Context) (bool, error) {
	err := sp.Ping(ctx)
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthCompressor, "%s", err)
	}
	return true, nil
}
