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

func TestControllerHandleAlbum(t *testing.T) {
	type give struct {
		err        error
		filenames  []string
		durationOn bool
		duration   string
	}
	type want struct {
		code int
		typ  string
		body string
	}
	tests := []struct {
		give
		want
	}{
		{
			give: give{
				err:        nil,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 201,
				typ:  "application/json; charset=utf-8",
				body: `{"album":{"id":"rRsAAAAAAAA"}}` + "\n",
			},
		},
		{
			give: give{
				err:        nil,
				filenames:  []string{"big.jpg", "big.jpg", "big.jpg"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":2,"msg":"body too large"}}` + "\n",
			},
		},
		{
			give: give{
				err:        nil,
				filenames:  []string{"alan.jpg"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":4,"msg":"not enough images"}}` + "\n",
			},
		},
		{
			give: give{
				err:        nil,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png", "alan.jpg"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":5,"msg":"too many images"}}` + "\n",
			},
		},
		{
			give: give{
				err:        nil,
				filenames:  []string{"alan.jpg", "john.bmp", "big.jpg"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":6,"msg":"image too large"}}` + "\n",
			},
		},
		{
			give: give{
				err:        nil,
				filenames:  []string{"alan.jpg", "john.bmp", "audio.ogg"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":7,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err:        nil,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: false,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":8,"msg":"duration not set"}}` + "\n",
			},
		},
		{
			give: give{
				err:        nil,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "",
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":9,"msg":"duration invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrTooManyRequests,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 429,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":1,"msg":"too many requests"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrBodyTooLarge,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":2,"msg":"body too large"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrWrongContentType,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":3,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrNotEnoughImages,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":4,"msg":"not enough images"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrTooManyImages,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "11h",
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":5,"msg":"too many images"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrImageTooLarge,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":6,"msg":"image too large"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrNotImage,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":7,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrDurationNotSet,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":8,"msg":"duration not set"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrDurationInvalid,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":9,"msg":"duration invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrAlbumNotFound,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":10,"msg":"album not found"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrTokenNotFound,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":12,"msg":"token not found"}}` + "\n",
			},
		},
		{
			give: give{
				err:        domain.ErrThirdPartyUnavailable,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":17,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err:        context.Canceled,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-1,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err:        context.DeadlineExceeded,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-2,"msg":"internal server error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(DefaultControllerConfig, serv)
			fn := contr.handleAlbum()
			w := httptest.NewRecorder()
			body := bytes.Buffer{}
			multi := multipart.NewWriter(&body)
			for _, filename := range tt.give.filenames {
				part, err := multi.CreateFormFile("images", filename)
				assert.NoError(t, err)
				b, err := os.ReadFile("../../testdata/" + filename)
				assert.NoError(t, err)
				_, err = part.Write(b)
				assert.NoError(t, err)
			}
			if tt.give.durationOn {
				err := multi.WriteField("duration", tt.give.duration)
				assert.NoError(t, err)
			}
			err := multi.Close()
			assert.NoError(t, err)
			r := httptest.NewRequest(http.MethodPost, "/api/albums/", &body)
			r.Header.Set("Content-Type", multi.FormDataContentType())
			fn(w, r, nil)
			AssertStatusCode(t, w, tt.want.code)
			AssertContentType(t, w, tt.want.typ)
			AssertBody(t, w, tt.want.body)
		})
	}
}

func TestControllerHandleStatus(t *testing.T) {
	type give struct {
		err error
	}
	type want struct {
		code int
		typ  string
		body string
	}
	tests := []struct {
		give
		want
	}{
		{
			give: give{
				err: nil,
			},
			want: want{
				code: 200,
				typ:  "application/json; charset=utf-8",
				body: `{"album":{"compression":{"progress":1}}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTooManyRequests,
			},
			want: want{
				code: 429,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":1,"msg":"too many requests"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBodyTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":2,"msg":"body too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrWrongContentType,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":3,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotEnoughImages,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":4,"msg":"not enough images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTooManyImages,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":5,"msg":"too many images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrImageTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":6,"msg":"image too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotImage,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":7,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationNotSet,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":8,"msg":"duration not set"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationInvalid,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":9,"msg":"duration invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrAlbumNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":10,"msg":"album not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTokenNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":12,"msg":"token not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrThirdPartyUnavailable,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":17,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.Canceled,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-1,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.DeadlineExceeded,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-2,"msg":"internal server error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(DefaultControllerConfig, serv)
			fn := contr.handleStatus()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/albums/rRsAAAAAAAA/status", http.NoBody)
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "rRsAAAAAAAA"}}
			fn(w, r, ps)
			AssertStatusCode(t, w, tt.want.code)
			AssertContentType(t, w, tt.want.typ)
			AssertBody(t, w, tt.want.body)
		})
	}
}

func TestControllerHandlePair(t *testing.T) {
	type give struct {
		err error
	}
	type want struct {
		code int
		typ  string
		body string
	}
	tests := []struct {
		give
		want
	}{
		{
			give: give{
				err: nil,
			},
			want: want{
				code: 200,
				typ:  "application/json; charset=utf-8",
				body: `{"album":{"img1":{"token":"f8cAAAAAAAA","src":"/aye-and-nay/albums/nkUAAAAAAAA/images/21EAAAAAAAA"},"img2":{"token":"iakAAAAAAAA","src":"/aye-and-nay/albums/nkUAAAAAAAA/images/K2IAAAAAAAA"}}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTooManyRequests,
			},
			want: want{
				code: 429,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":1,"msg":"too many requests"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBodyTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":2,"msg":"body too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrWrongContentType,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":3,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotEnoughImages,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":4,"msg":"not enough images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTooManyImages,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":5,"msg":"too many images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrImageTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":6,"msg":"image too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotImage,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":7,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationNotSet,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":8,"msg":"duration not set"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationInvalid,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":9,"msg":"duration invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrAlbumNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":10,"msg":"album not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTokenNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":12,"msg":"token not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrThirdPartyUnavailable,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":17,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.Canceled,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-1,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.DeadlineExceeded,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-2,"msg":"internal server error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(DefaultControllerConfig, serv)
			fn := contr.handlePair()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/albums/nkUAAAAAAAA/", http.NoBody)
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "nkUAAAAAAAA"}}
			fn(w, r, ps)
			AssertStatusCode(t, w, tt.want.code)
			AssertContentType(t, w, tt.want.typ)
			AssertBody(t, w, tt.want.body)
		})
	}
}

func TestControllerHandleImage(t *testing.T) {
	body, _ := io.ReadAll(Png())
	type give struct {
		err error
	}
	type want struct {
		code int
		typ  string
		body string
	}
	tests := []struct {
		give
		want
	}{
		{
			give: give{
				err: nil,
			},
			want: want{
				code: 200,
				typ:  "image/png",
				body: string(body),
			},
		},
		{
			give: give{
				err: domain.ErrTooManyRequests,
			},
			want: want{
				code: 429,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":1,"msg":"too many requests"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBodyTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":2,"msg":"body too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrWrongContentType,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":3,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotEnoughImages,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":4,"msg":"not enough images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTooManyImages,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":5,"msg":"too many images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrImageTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":6,"msg":"image too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotImage,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":7,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationNotSet,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":8,"msg":"duration not set"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationInvalid,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":9,"msg":"duration invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrAlbumNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":10,"msg":"album not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTokenNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":12,"msg":"token not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrThirdPartyUnavailable,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":17,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.Canceled,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-1,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.DeadlineExceeded,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-2,"msg":"internal server error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(DefaultControllerConfig, serv)
			fn := contr.handleImage()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/images/8v7AAAAAAAA/", http.NoBody)
			ps := httprouter.Params{httprouter.Param{Key: "token", Value: "8v7AAAAAAAA"}}
			fn(w, r, ps)
			AssertStatusCode(t, w, tt.want.code)
			AssertContentType(t, w, tt.want.typ)
			AssertBody(t, w, tt.want.body)
		})
	}
}

func TestControllerHandleVote(t *testing.T) {
	type give struct {
		err error
	}
	type want struct {
		code int
		typ  string
		body string
	}
	tests := []struct {
		give
		want
	}{
		{
			give: give{
				err: nil,
			},
			want: want{
				code: 200,
				typ:  "",
				body: ``,
			},
		},
		{
			give: give{
				err: domain.ErrTooManyRequests,
			},
			want: want{
				code: 429,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":1,"msg":"too many requests"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBodyTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":2,"msg":"body too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrWrongContentType,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":3,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotEnoughImages,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":4,"msg":"not enough images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTooManyImages,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":5,"msg":"too many images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrImageTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":6,"msg":"image too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotImage,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":7,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationNotSet,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":8,"msg":"duration not set"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationInvalid,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":9,"msg":"duration invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrAlbumNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":10,"msg":"album not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTokenNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":12,"msg":"token not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrThirdPartyUnavailable,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":17,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.Canceled,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-1,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.DeadlineExceeded,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-2,"msg":"internal server error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(DefaultControllerConfig, serv)
			fn := contr.handleVote()
			w := httptest.NewRecorder()
			json := strings.NewReader(`{"album":{"imgFrom":{"token":"fYIAAAAAAAA"},"imgTo":{"token":"foIAAAAAAAA"}}}`)
			r := httptest.NewRequest(http.MethodPatch, "/api/albums/fIIAAAAAAAA/", json)
			r.Header.Set("Content-Type", "application/json; charset=utf-8")
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "fIIAAAAAAAA"}}
			fn(w, r, ps)
			AssertStatusCode(t, w, tt.want.code)
			AssertContentType(t, w, tt.want.typ)
			AssertBody(t, w, tt.want.body)
		})
	}
}

func TestControllerHandleTop(t *testing.T) {
	type give struct {
		err error
	}
	type want struct {
		code int
		typ  string
		body string
	}
	tests := []struct {
		give
		want
	}{
		{
			give: give{
				err: nil,
			},
			want: want{
				code: 200,
				typ:  "application/json; charset=utf-8",
				body: `{"album":{"images":[{"src":"/aye-and-nay/albums/byYAAAAAAAA/images/yFwAAAAAAAA","rating":0.5},{"src":"/aye-and-nay/albums/byYAAAAAAAA/images/jVgAAAAAAAA","rating":0.5}]}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTooManyRequests,
			},
			want: want{
				code: 429,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":1,"msg":"too many requests"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBodyTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":2,"msg":"body too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrWrongContentType,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":3,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotEnoughImages,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":4,"msg":"not enough images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTooManyImages,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":5,"msg":"too many images"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrImageTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":6,"msg":"image too large"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrNotImage,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":7,"msg":"unsupported media type"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationNotSet,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":8,"msg":"duration not set"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrDurationInvalid,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":9,"msg":"duration invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrAlbumNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":10,"msg":"album not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrTokenNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":12,"msg":"token not found"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrThirdPartyUnavailable,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":17,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.Canceled,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-1,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.DeadlineExceeded,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":-2,"msg":"internal server error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(DefaultControllerConfig, serv)
			fn := contr.handleTop()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/albums/byYAAAAAAAA/top/", http.NoBody)
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "byYAAAAAAAA"}}
			fn(w, r, ps)
			AssertStatusCode(t, w, tt.want.code)
			AssertContentType(t, w, tt.want.typ)
			AssertBody(t, w, tt.want.body)
		})
	}
}

func TestControllerHandleHealth(t *testing.T) {
	type give struct {
		err error
	}
	type want struct {
		code int
		typ  string
		body string
	}
	tests := []struct {
		give
		want
	}{
		{
			give: give{
				err: nil,
			},
			want: want{
				code: 200,
				typ:  "",
				body: ``,
			},
		},
		{
			give: give{
				err: domain.ErrBadHealthCompressor,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":18,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBadHealthStorage,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":19,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBadHealthDatabase,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":20,"msg":"internal server error"}}` + "\n",
			},
		},
		{
			give: give{
				err: domain.ErrBadHealthCache,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"code":21,"msg":"internal server error"}}` + "\n",
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
			ps := httprouter.Params{}
			fn(w, r, ps)
			AssertStatusCode(t, w, tt.want.code)
			AssertContentType(t, w, tt.want.typ)
			AssertBody(t, w, tt.want.body)
		})
	}
}
