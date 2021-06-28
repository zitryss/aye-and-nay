package http

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/domain/service"
	_ "github.com/zitryss/aye-and-nay/internal/config"
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
				body: `{"error":{"msg":"Body Too Large"}}` + "\n",
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
				body: `{"error":{"msg":"Not Enough Images"}}` + "\n",
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
				body: `{"error":{"msg":"Too Many Images"}}` + "\n",
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
				body: `{"error":{"msg":"Image Too Large"}}` + "\n",
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
				body: `{"error":{"msg":"Unsupported Image Format"}}` + "\n",
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
				body: `{"error":{"msg":"Duration Not Set"}}` + "\n",
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
				body: `{"error":{"msg":"Duration Invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrTooManyRequests,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 429,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Too Many Requests"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrBodyTooLarge,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Body Too Large"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrWrongContentType,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Unsupported Content Type"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrNotEnoughImages,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Not Enough Images"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrTooManyImages,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "11h",
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Too Many Images"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrImageTooLarge,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Image Too Large"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrNotImage,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Unsupported Image Format"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrDurationNotSet,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Duration Not Set"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrDurationInvalid,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Duration Invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrAlbumNotFound,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Album Not Found"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrTokenNotFound,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Token Not Found"}}` + "\n",
			},
		},
		{
			give: give{
				err:        model.ErrThirdPartyUnavailable,
				filenames:  []string{"alan.jpg", "john.bmp", "dennis.png"},
				durationOn: true,
				duration:   "1h",
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
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
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
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
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(serv)
			fn := contr.handleAlbum()
			w := httptest.NewRecorder()
			body := bytes.Buffer{}
			multi := multipart.NewWriter(&body)
			for _, filename := range tt.give.filenames {
				part, err := multi.CreateFormFile("images", filename)
				if err != nil {
					t.Error(err)
				}
				b, err := os.ReadFile("../../testdata/" + filename)
				if err != nil {
					t.Error(err)
				}
				_, err = part.Write(b)
				if err != nil {
					t.Error(err)
				}
			}
			if tt.give.durationOn {
				err := multi.WriteField("duration", tt.give.duration)
				if err != nil {
					t.Error(err)
				}
			}
			err := multi.Close()
			if err != nil {
				t.Error(err)
			}
			r := httptest.NewRequest("POST", "/api/albums/", &body)
			r.Header.Set("Content-Type", multi.FormDataContentType())
			fn(w, r, nil)
			CheckStatusCode(t, w, tt.want.code)
			CheckContentType(t, w, tt.want.typ)
			CheckBody(t, w, tt.want.body)
		})
	}
}

func TestControllerHandleReady(t *testing.T) {
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
				body: `{"album":{"progress":1}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrTooManyRequests,
			},
			want: want{
				code: 429,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Too Many Requests"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrBodyTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Body Too Large"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrWrongContentType,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Unsupported Content Type"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrNotEnoughImages,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Not Enough Images"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrTooManyImages,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Too Many Images"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrImageTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Image Too Large"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrNotImage,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Unsupported Image Format"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrDurationNotSet,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Duration Not Set"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrDurationInvalid,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Duration Invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrAlbumNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Album Not Found"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrTokenNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Token Not Found"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrThirdPartyUnavailable,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.Canceled,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.DeadlineExceeded,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(serv)
			fn := contr.handleReady()
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/api/albums/rRsAAAAAAAA/ready", nil)
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "rRsAAAAAAAA"}}
			fn(w, r, ps)
			CheckStatusCode(t, w, tt.want.code)
			CheckContentType(t, w, tt.want.typ)
			CheckBody(t, w, tt.want.body)
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
				err: model.ErrTooManyRequests,
			},
			want: want{
				code: 429,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Too Many Requests"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrBodyTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Body Too Large"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrWrongContentType,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Unsupported Content Type"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrNotEnoughImages,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Not Enough Images"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrTooManyImages,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Too Many Images"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrImageTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Image Too Large"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrNotImage,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Unsupported Image Format"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrDurationNotSet,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Duration Not Set"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrDurationInvalid,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Duration Invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrAlbumNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Album Not Found"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrTokenNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Token Not Found"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrThirdPartyUnavailable,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.Canceled,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.DeadlineExceeded,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(serv)
			fn := contr.handlePair()
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/api/albums/nkUAAAAAAAA/", nil)
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "nkUAAAAAAAA"}}
			fn(w, r, ps)
			CheckStatusCode(t, w, tt.want.code)
			CheckContentType(t, w, tt.want.typ)
			CheckBody(t, w, tt.want.body)
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
				err: model.ErrTooManyRequests,
			},
			want: want{
				code: 429,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Too Many Requests"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrBodyTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Body Too Large"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrWrongContentType,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Unsupported Content Type"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrNotEnoughImages,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Not Enough Images"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrTooManyImages,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Too Many Images"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrImageTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Image Too Large"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrNotImage,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Unsupported Image Format"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrDurationNotSet,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Duration Not Set"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrDurationInvalid,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Duration Invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrAlbumNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Album Not Found"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrTokenNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Token Not Found"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrThirdPartyUnavailable,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.Canceled,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.DeadlineExceeded,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(serv)
			fn := contr.handleVote()
			w := httptest.NewRecorder()
			json := strings.NewReader(`{"album":{"imgFrom":{"token":"fYIAAAAAAAA"},"imgTo":{"token":"foIAAAAAAAA"}}}`)
			r := httptest.NewRequest("PATCH", "/api/albums/fIIAAAAAAAA/", json)
			r.Header.Set("Content-Type", "application/json; charset=utf-8")
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "fIIAAAAAAAA"}}
			fn(w, r, ps)
			CheckStatusCode(t, w, tt.want.code)
			CheckContentType(t, w, tt.want.typ)
			CheckBody(t, w, tt.want.body)
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
				err: model.ErrTooManyRequests,
			},
			want: want{
				code: 429,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Too Many Requests"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrBodyTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Body Too Large"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrWrongContentType,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Unsupported Content Type"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrNotEnoughImages,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Not Enough Images"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrTooManyImages,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Too Many Images"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrImageTooLarge,
			},
			want: want{
				code: 413,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Image Too Large"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrNotImage,
			},
			want: want{
				code: 415,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Unsupported Image Format"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrDurationNotSet,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Duration Not Set"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrDurationInvalid,
			},
			want: want{
				code: 400,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Duration Invalid"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrAlbumNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Album Not Found"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrTokenNotFound,
			},
			want: want{
				code: 404,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Token Not Found"}}` + "\n",
			},
		},
		{
			give: give{
				err: model.ErrThirdPartyUnavailable,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.Canceled,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
		{
			give: give{
				err: context.DeadlineExceeded,
			},
			want: want{
				code: 500,
				typ:  "application/json; charset=utf-8",
				body: `{"error":{"msg":"Internal Server Error"}}` + "\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			serv := service.NewMock(tt.give.err)
			contr := newController(serv)
			fn := contr.handleTop()
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/api/albums/byYAAAAAAAA/top/", nil)
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "byYAAAAAAAA"}}
			fn(w, r, ps)
			CheckStatusCode(t, w, tt.want.code)
			CheckContentType(t, w, tt.want.typ)
			CheckBody(t, w, tt.want.body)
		})
	}
}
