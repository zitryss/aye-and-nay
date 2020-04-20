package http

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/julienschmidt/httprouter"

	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/rand"
)

func TestHtmlHandleAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("y9FYFBmQyvdzYBY7", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		html, err := newHtml(&serv)
		if err != nil {
			t.Error(err)
		}
		fn := html.handleAlbum()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		fn(w, r, nil)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "text/html; charset=utf-8")
	})
}

func TestHtmlHandlePair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "BLN3fureNCB7w44Z"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("9zYp2QUKJx8nsuKX", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
		html, err := newHtml(&serv)
		if err != nil {
			t.Error(err)
		}
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
		err = multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		fn = html.handlePair()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/albums/BLN3fureNCB7w44Z1/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "BLN3fureNCB7w44Z1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "text/html; charset=utf-8")
	})
	t.Run("Negative", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("NDUAttuKk2NaJAwq", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		html, err := newHtml(&serv)
		if err != nil {
			t.Error(err)
		}
		fn := html.handlePair()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/albums/3gVXkrERt9M6eeRW/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "3gVXkrERt9M6eeRW"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 404)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Page Not Found`+"\n")
	})
}

func TestHtmlHandleTop(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "e54aXenm2z4gyNFy"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("zfAL4ZAqQ84kQSHf", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
		html, err := newHtml(&serv)
		if err != nil {
			t.Error(err)
		}
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
		err = multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		fn = html.handleTop()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/albums/e54aXenm2z4gyNFy1/top/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "e54aXenm2z4gyNFy1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "text/html; charset=utf-8")
	})
	t.Run("Negative", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("WDsKDNfNM4Bt7UuB", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		html, err := newHtml(&serv)
		if err != nil {
			t.Error(err)
		}
		fn := html.handleTop()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/albums/pBLBLqq7Pu7jDhJ5/top/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "pBLBLqq7Pu7jDhJ5"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 404)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Page Not Found`+"\n")
	})
}
