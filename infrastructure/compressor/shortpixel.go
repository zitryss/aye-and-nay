package compressor

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/zitryss/aye-and-nay/domain/model"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

func NewShortPixel() shortpixel {
	conf := newShortPixelConfig()
	return shortpixel{conf}
}

type shortpixel struct {
	conf shortPixelConfig
}

func (sp *shortpixel) Ping() error {
	img := model.Image{Id: "test", B: Png()}
	err := sp.upload(context.Background(), &img)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (sp *shortpixel) Compress(ctx context.Context, imgs []model.Image) error {
	for i := range imgs {
		img := &imgs[i]
		err := sp.upload(ctx, img)
		if err != nil {
			return errors.Wrap(err)
		}
		err = sp.download(ctx, img)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	}
	return nil
}

func (sp *shortpixel) upload(ctx context.Context, img *model.Image) error {
	body := bytes.Buffer{}
	multi := multipart.NewWriter(&body)
	part, err := multi.CreateFormField("key")
	if err != nil {
		return errors.Wrap(err)
	}
	_, err = io.WriteString(part, sp.conf.apiKey)
	if err != nil {
		return errors.Wrap(err)
	}
	part, err = multi.CreateFormField("lossy")
	if err != nil {
		return errors.Wrap(err)
	}
	_, err = io.WriteString(part, "1")
	if err != nil {
		return errors.Wrap(err)
	}
	part, err = multi.CreateFormField("wait")
	if err != nil {
		return errors.Wrap(err)
	}
	_, err = io.WriteString(part, sp.conf.wait)
	if err != nil {
		return errors.Wrap(err)
	}
	part, err = multi.CreateFormField("convertto")
	if err != nil {
		return errors.Wrap(err)
	}
	_, err = io.WriteString(part, "png")
	if err != nil {
		return errors.Wrap(err)
	}
	part, err = multi.CreateFormField("file_paths")
	if err != nil {
		return errors.Wrap(err)
	}
	_, err = io.WriteString(part, `{"file": ""}`)
	if err != nil {
		return errors.Wrap(err)
	}
	part, err = multi.CreateFormFile("file", img.Id)
	if err != nil {
		return errors.Wrap(err)
	}
	_, err = part.Write(img.B)
	if err != nil {
		return errors.Wrap(err)
	}
	err = multi.Close()
	if err != nil {
		return errors.Wrap(err)
	}
	c := http.Client{Timeout: sp.conf.uploadTimeout}
	req, err := http.NewRequestWithContext(ctx, "POST", sp.conf.url, &body)
	if err != nil {
		return errors.Wrap(err)
	}
	req.Header.Set("Content-Type", multi.FormDataContentType())
	resp := (*http.Response)(nil)
	err = retry.Do(sp.conf.times, sp.conf.pause, func() error {
		resp, err = c.Do(req)
		if err != nil {
			return errors.Wrapf(model.ErrThirdPartyUnavailable, "%s", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			_, _ = io.Copy(ioutil.Discard, resp.Body)
			_ = resp.Body.Close()
			return errors.Wrapf(model.ErrThirdPartyUnavailable, "%s", err)
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err)
	}
	buf := bytes.Buffer{}
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
		return errors.Wrap(err)
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
		OriginalUrl string
		LossyUrl    string
	}{}
	err = json.NewDecoder(&buf).Decode(&response)
	if err != nil {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
		return errors.Wrapf(model.ErrThirdPartyUnavailable, "%s", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return errors.Wrap(err)
	}
	switch response.Status.Code {
	case "1":
		img.Src = response.OriginalUrl
		err = sp.repeat(ctx, img)
		if err != nil {
			return errors.Wrap(err)
		}
	case "2":
		img.Src = response.LossyUrl
	case -201.0, -202.0:
		return errors.Wrap(model.ErrNotImage)
	default:
		return errors.Wrapf(model.ErrThirdPartyUnavailable, "status code %v: message %q", response.Status.Code, response.Status.Message)
	}
	return nil
}

func (sp *shortpixel) repeat(ctx context.Context, img *model.Image) error {
	time.Sleep(sp.conf.repeatIn)
	body := bytes.Buffer{}
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
		Urllist:   []string{img.Src},
	}
	err := json.NewEncoder(&body).Encode(request)
	if err != nil {
		return errors.Wrap(err)
	}
	c := http.Client{Timeout: sp.conf.uploadTimeout}
	req, err := http.NewRequestWithContext(ctx, "POST", sp.conf.url2, &body)
	if err != nil {
		return errors.Wrap(err)
	}
	resp := (*http.Response)(nil)
	err = retry.Do(sp.conf.times, sp.conf.pause, func() error {
		resp, err = c.Do(req)
		if err != nil {
			return errors.Wrapf(model.ErrThirdPartyUnavailable, "%s", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			_, _ = io.Copy(ioutil.Discard, resp.Body)
			_ = resp.Body.Close()
			return errors.Wrapf(model.ErrThirdPartyUnavailable, "%s", err)
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err)
	}
	buf := bytes.Buffer{}
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
		return errors.Wrap(err)
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
	err = json.NewDecoder(&buf).Decode(&response)
	if err != nil {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
		return errors.Wrapf(model.ErrThirdPartyUnavailable, "%s", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return errors.Wrap(err)
	}
	switch response.Status.Code {
	case "1":
		return errors.Wrapf(model.ErrThirdPartyUnavailable, "status code %v: message %q", response.Status.Code, response.Status.Message)
	case "2":
		img.Src = response.LossyUrl
	default:
		return errors.Wrapf(model.ErrThirdPartyUnavailable, "status code %v: message %q", response.Status.Code, response.Status.Message)
	}
	return nil
}

func (sp *shortpixel) download(ctx context.Context, img *model.Image) error {
	c := http.Client{Timeout: sp.conf.downloadTimeout}
	req, err := http.NewRequestWithContext(ctx, "GET", img.Src, nil)
	if err != nil {
		return errors.Wrap(err)
	}
	resp := (*http.Response)(nil)
	err = retry.Do(sp.conf.times, sp.conf.pause, func() error {
		resp, err = c.Do(req)
		if err != nil {
			return errors.Wrapf(model.ErrThirdPartyUnavailable, "%s", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			_, _ = io.Copy(ioutil.Discard, resp.Body)
			_ = resp.Body.Close()
			return errors.Wrapf(model.ErrThirdPartyUnavailable, "%s", err)
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err)
	}
	n, err := io.ReadFull(resp.Body, img.B)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
		return errors.Wrap(err)
	}
	err = resp.Body.Close()
	if err != nil {
		return errors.Wrap(err)
	}
	img.B = img.B[:n]
	img.Src = ""
	return nil
}
