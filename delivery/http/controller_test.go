//go:build unit

package http

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/service"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestControllerHandle(t *testing.T) {
	type give struct {
		handle  func() httprouter.Handle
		method  string
		target  string
		reqBody io.Reader
		headers map[string]string
		params  []httprouter.Param
	}
	type want struct {
		code     int
		typ      string
		respBody string
	}
	contr := controller{}
	payload := content{}
	tests := []struct {
		give
		want
	}{
		{
			give: give{
				handle:  contr.handleAlbum,
				method:  http.MethodPost,
				target:  "/api/albums/",
				reqBody: payload.body(t, []string{"alan.jpg", "john.bmp", "dennis.png"}, true, "1h"),
				headers: map[string]string{"Content-Type": payload.boundary},
			},
			want: want{
				code:     http.StatusCreated,
				typ:      "application/json; charset=utf-8",
				respBody: `{"album":{"id":"rRsAAAAAAAA"}}` + "\n",
			},
		},
		{
			give: give{
				handle:  contr.handleAlbum,
				method:  http.MethodPost,
				target:  "/api/albums/",
				reqBody: payload.body(t, []string{"big.jpg", "big.jpg", "big.jpg"}, true, "1h"),
				headers: map[string]string{"Content-Type": payload.boundary},
			},
			want: want{
				code:     http.StatusRequestEntityTooLarge,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":2,"msg":"body too large"}}` + "\n",
			},
		},
		{
			give: give{
				handle:  contr.handleAlbum,
				method:  http.MethodPost,
				target:  "/api/albums/",
				reqBody: payload.body(t, []string{"alan.jpg"}, true, "1h"),
				headers: map[string]string{"Content-Type": payload.boundary},
			},
			want: want{
				code:     http.StatusBadRequest,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":4,"msg":"not enough images"}}` + "\n",
			},
		},
		{
			give: give{
				handle:  contr.handleAlbum,
				method:  http.MethodPost,
				target:  "/api/albums/",
				reqBody: payload.body(t, []string{"alan.jpg", "john.bmp", "dennis.png", "alan.jpg"}, true, "1h"),
				headers: map[string]string{"Content-Type": payload.boundary},
			},
			want: want{
				code:     http.StatusRequestEntityTooLarge,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":5,"msg":"too many images"}}` + "\n",
			},
		},
		{
			give: give{
				handle:  contr.handleAlbum,
				method:  http.MethodPost,
				target:  "/api/albums/",
				reqBody: payload.body(t, []string{"alan.jpg", "john.bmp", "big.jpg"}, true, "1h"),
				headers: map[string]string{"Content-Type": payload.boundary},
			},
			want: want{
				code:     http.StatusRequestEntityTooLarge,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":6,"msg":"image too large"}}` + "\n",
			},
		},
		{
			give: give{
				handle:  contr.handleAlbum,
				method:  http.MethodPost,
				target:  "/api/albums/",
				reqBody: payload.body(t, []string{"alan.jpg", "john.bmp", "audio.ogg"}, true, "1h"),
				headers: map[string]string{"Content-Type": payload.boundary},
			},
			want: want{
				code:     http.StatusUnsupportedMediaType,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":7,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				handle:  contr.handleAlbum,
				method:  http.MethodPost,
				target:  "/api/albums/",
				reqBody: payload.body(t, []string{"alan.jpg", "john.bmp", "dennis.png"}, false, ""),
				headers: map[string]string{"Content-Type": payload.boundary},
			},
			want: want{
				code:     http.StatusBadRequest,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":8,"msg":"duration not set"}}` + "\n",
			},
		},
		{
			give: give{
				handle:  contr.handleAlbum,
				method:  http.MethodPost,
				target:  "/api/albums/",
				reqBody: payload.body(t, []string{"alan.jpg", "john.bmp", "dennis.png"}, true, ""),
				headers: map[string]string{"Content-Type": payload.boundary},
			},
			want: want{
				code:     http.StatusBadRequest,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":9,"msg":"duration invalid"}}` + "\n",
			},
		},
		{
			give: give{
				handle: contr.handleStatus,
				method: http.MethodGet,
				target: "/api/albums/rRsAAAAAAAA/status",
				params: httprouter.Params{httprouter.Param{Key: "album", Value: "rRsAAAAAAAA"}},
			},
			want: want{
				code:     http.StatusOK,
				typ:      "application/json; charset=utf-8",
				respBody: `{"album":{"compression":{"progress":1}}}` + "\n",
			},
		},
		{
			give: give{
				handle: contr.handlePair,
				method: http.MethodGet,
				target: "/api/albums/nkUAAAAAAAA/",
				params: httprouter.Params{httprouter.Param{Key: "album", Value: "nkUAAAAAAAA"}},
			},
			want: want{
				code:     http.StatusOK,
				typ:      "application/json; charset=utf-8",
				respBody: `{"album":{"img1":{"token":"f8cAAAAAAAA","src":"/aye-and-nay/albums/nkUAAAAAAAA/images/21EAAAAAAAA"},"img2":{"token":"iakAAAAAAAA","src":"/aye-and-nay/albums/nkUAAAAAAAA/images/K2IAAAAAAAA"}}}` + "\n",
			},
		},
		{
			give: give{
				handle: contr.handleImage,
				method: http.MethodGet,
				target: "/api/images/8v7AAAAAAAA/",
				params: httprouter.Params{httprouter.Param{Key: "token", Value: "8v7AAAAAAAA"}},
			},
			want: want{
				code:     http.StatusOK,
				typ:      "image/png",
				respBody: png(),
			},
		},
		{
			give: give{
				handle:  contr.handleVote,
				method:  http.MethodPatch,
				target:  "/api/albums/fIIAAAAAAAA/",
				reqBody: strings.NewReader(`{"album":{"imgFrom":{"token":"fYIAAAAAAAA"},"imgTo":{"token":"foIAAAAAAAA"}}}`),
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				params:  httprouter.Params{httprouter.Param{Key: "album", Value: "fIIAAAAAAAA"}},
			},
			want: want{
				code:     http.StatusOK,
				typ:      "",
				respBody: ``,
			},
		},
		{
			give: give{
				handle: contr.handleTop,
				method: http.MethodGet,
				target: "/api/albums/byYAAAAAAAA/top/",
				params: httprouter.Params{httprouter.Param{Key: "album", Value: "byYAAAAAAAA"}},
			},
			want: want{
				code:     http.StatusOK,
				typ:      "application/json; charset=utf-8",
				respBody: `{"album":{"images":[{"src":"/aye-and-nay/albums/byYAAAAAAAA/images/yFwAAAAAAAA","rating":0.5},{"src":"/aye-and-nay/albums/byYAAAAAAAA/images/jVgAAAAAAAA","rating":0.5}]}}` + "\n",
			},
		},
		{
			give: give{
				handle: contr.handleHealth,
				method: http.MethodGet,
				target: "/api/health/",
			},
			want: want{
				code:     http.StatusOK,
				typ:      "",
				respBody: ``,
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			err := error(nil)
			serv := service.NewMock(err)
			contr = newController(DefaultControllerConfig, serv)
			fn := tt.give.handle()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.give.method, tt.give.target, tt.give.reqBody)
			for k, v := range tt.give.headers {
				r.Header.Set(k, v)
			}
			fn(w, r, tt.give.params)
			AssertStatusCode(t, w, tt.want.code)
			AssertContentType(t, w, tt.want.typ)
			AssertBody(t, w, tt.want.respBody)
		})
	}
}

type content struct {
	boundary string
}

func (c *content) body(t *testing.T, filenames []string, durationOn bool, duration string) io.Reader {
	t.Helper()
	body := bytes.Buffer{}
	multi := multipart.NewWriter(&body)
	for _, filename := range filenames {
		part, err := multi.CreateFormFile("images", filename)
		assert.NoError(t, err)
		b, err := os.ReadFile("../../testdata/" + filename)
		assert.NoError(t, err)
		_, err = part.Write(b)
		assert.NoError(t, err)
	}
	if durationOn {
		err := multi.WriteField("duration", duration)
		assert.NoError(t, err)
	}
	err := multi.Close()
	assert.NoError(t, err)
	c.boundary = multi.FormDataContentType()
	return &body
}

func png() string {
	body, _ := io.ReadAll(Png())
	return string(body)
}

func TestControllerError(t *testing.T) {
	type give struct {
		err error
	}
	type want struct {
		code     int
		typ      string
		respBody string
	}
	tests := []struct {
		give
		want
	}{
		{
			give: give{
				err: domain.ErrTooManyRequests,
			},
			want: want{
				code:     http.StatusTooManyRequests,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":1,"msg":"too many requests"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBodyTooLarge,
			},
			want: want{
				code:     http.StatusRequestEntityTooLarge,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":2,"msg":"body too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrWrongContentType,
			},
			want: want{
				code:     http.StatusUnsupportedMediaType,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":3,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotEnoughImages,
			},
			want: want{
				code:     http.StatusBadRequest,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":4,"msg":"not enough images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTooManyImages,
			},
			want: want{
				code:     http.StatusRequestEntityTooLarge,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":5,"msg":"too many images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrImageTooLarge,
			},
			want: want{
				code:     http.StatusRequestEntityTooLarge,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":6,"msg":"image too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotImage,
			},
			want: want{
				code:     http.StatusUnsupportedMediaType,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":7,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationNotSet,
			},
			want: want{
				code:     http.StatusBadRequest,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":8,"msg":"duration not set"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationInvalid,
			},
			want: want{
				code:     http.StatusBadRequest,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":9,"msg":"duration invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrAlbumNotFound,
			},
			want: want{
				code:     http.StatusNotFound,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":10,"msg":"album not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrPairNotFound,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":11,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTokenNotFound,
			},
			want: want{
				code:     http.StatusNotFound,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":12,"msg":"token not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrImageNotFound,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":13,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrAlbumAlreadyExists,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":14,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTokenAlreadyExists,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":15,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrUnsupportedMediaType,
			},
			want: want{
				code:     http.StatusUnsupportedMediaType,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":16,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrThirdPartyUnavailable,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":17,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBadHealthCompressor,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":18,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBadHealthStorage,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":19,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBadHealthDatabase,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":20,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBadHealthCache,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":21,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrUnknown,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":22,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.Canceled,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":-1,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.DeadlineExceeded,
			},
			want: want{
				code:     http.StatusInternalServerError,
				typ:      "application/json; charset=utf-8",
				respBody: `{"error":{"code":-2,"msg":"internal server error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(DefaultControllerConfig, serv)
			fn := contr.handleHealth()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/health/", http.NoBody)
			fn(w, r, nil)
			AssertStatusCode(t, w, tt.want.code)
			AssertContentType(t, w, tt.want.typ)
			AssertBody(t, w, tt.want.respBody)
		})
	}
}
