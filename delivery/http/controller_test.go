package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
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
				code: 201,
				typ:  "application/json; charset=utf-8",
				body: `{"album":{"id":"N2fxX5zbDh8RJQvx1"}}` + "\n",
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
			fn := contr.handleAlbum()
			w := httptest.NewRecorder()
			body := bytes.Buffer{}
			multi := multipart.NewWriter(&body)
			for _, filename := range []string{"alan.jpg", "john.bmp", "dennis.png"} {
				part, err := multi.CreateFormFile("images", filename)
				if err != nil {
					t.Error(err)
				}
				b, err := ioutil.ReadFile("../../testdata/" + filename)
				if err != nil {
					t.Error(err)
				}
				_, err = part.Write(b)
				if err != nil {
					t.Error(err)
				}
			}
			err := multi.WriteField("duration", "1H")
			if err != nil {
				t.Error(err)
			}
			err = multi.Close()
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
			r := httptest.NewRequest("GET", "/api/albums/N2fxX5zbDh8RJQvx1/ready", nil)
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "N2fxX5zbDh8RJQvx1"}}
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
				body: `{"album":{"img1":{"token":"DfsXRkDxVeH2xmme5","src":"/aye-and-nay/albums/DfsXRkDxVeH2xmme1/images/DfsXRkDxVeH2xmme2"},"img2":{"token":"DfsXRkDxVeH2xmme6","src":"/aye-and-nay/albums/DfsXRkDxVeH2xmme1/images/DfsXRkDxVeH2xmme3"}}}` + "\n",
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
			r := httptest.NewRequest("GET", "/api/albums/DfsXRkDxVeH2xmme1/", nil)
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "DfsXRkDxVeH2xmme1"}}
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
			json := strings.NewReader(`{"album":{"imgFrom":{"token":"MvdZUxVgPD5p6JTa5"},"imgTo":{"token":"MvdZUxVgPD5p6JTa6"}}}`)
			r := httptest.NewRequest("PATCH", "/api/albums/MvdZUxVgPD5p6JTa1/", json)
			r.Header.Set("Content-Type", "application/json; charset=utf-8")
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "MvdZUxVgPD5p6JTa1"}}
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
				body: `{"album":{"images":[{"src":"/aye-and-nay/albums/bYCppY8q6qjvXjMZ1/images/bYCppY8q6qjvXjMZ2","rating":0.5},{"src":"/aye-and-nay/albums/bYCppY8q6qjvXjMZ1/images/bYCppY8q6qjvXjMZ3","rating":0.5}]}}` + "\n",
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
			r := httptest.NewRequest("GET", "/api/albums/bYCppY8q6qjvXjMZ1/top/", nil)
			ps := httprouter.Params{httprouter.Param{Key: "album", Value: "bYCppY8q6qjvXjMZ1"}}
			fn(w, r, ps)
			CheckStatusCode(t, w, tt.want.code)
			CheckContentType(t, w, tt.want.typ)
			CheckBody(t, w, tt.want.body)
		})
	}
}
