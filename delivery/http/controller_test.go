package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/rand"
)

func TestControllerHandleAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "N2fxX5zbDh8RJQvx"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("HVyMn8HuDa8rdkyr", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
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
		err := multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		CheckStatusCode(t, w, 201)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"album":{"id":"N2fxX5zbDh8RJQvx1"}}`+"\n")
	})
	t.Run("Negative1", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("mEdFrvE3549LDFzx", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
		fn := contr.handleAlbum()
		w := httptest.NewRecorder()
		body := bytes.Buffer{}
		multi := multipart.NewWriter(&body)
		for _, filename := range []string{"alan.jpg"} {
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
		err := multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		CheckStatusCode(t, w, 400)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Not Enough Images`+"\n")
	})
	t.Run("Negative2", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("xQWGJjTtetde2DdB", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
		fn := contr.handleAlbum()
		w := httptest.NewRecorder()
		body := bytes.Buffer{}
		multi := multipart.NewWriter(&body)
		for _, filename := range []string{"alan.jpg", "john.bmp", "dennis.png", "tim.gif"} {
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
		err := multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		CheckStatusCode(t, w, 413)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Request Entity Too Large`+"\n")
	})
	t.Run("Negative3", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("eQupRzAY56Qp5E4U", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
		fn := contr.handleAlbum()
		w := httptest.NewRecorder()
		body := bytes.Buffer{}
		multi := multipart.NewWriter(&body)
		for _, filename := range []string{"alan.jpg", "john.bmp", "linus.jpg"} {
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
		err := multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		CheckStatusCode(t, w, 413)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Request Entity Too Large`+"\n")
	})
	t.Run("Negative4", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("Au4DBRQhyEJV99wh", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
		fn := contr.handleAlbum()
		w := httptest.NewRecorder()
		body := bytes.Buffer{}
		multi := multipart.NewWriter(&body)
		for _, filename := range []string{"alan.jpg", "john.bmp", "neil.ogg"} {
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
		err := multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		CheckStatusCode(t, w, 415)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Unsupported Media Type`+"\n")
	})
	t.Run("Negative5", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "Ugr5UMNg2R6DKg43"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		comp := compressor.NewFail()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("UcNgY7Acvep8XqCc", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
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
		err := multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		CheckStatusCode(t, w, 500)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Internal Server Error`+"\n")
		w = httptest.NewRecorder()
		body = bytes.Buffer{}
		multi = multipart.NewWriter(&body)
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
		r = httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		CheckStatusCode(t, w, 201)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"album":{"id":"Ugr5UMNg2R6DKg435"}}`+"\n")
	})
}

func TestControllerHandlePair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "DfsXRkDxVeH2xmme"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		rand.Shuffle = func(n int, swap func(i int, j int)) {
		}
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("93P3AU2V6RMcFND4", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
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
		err := multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		fn = contr.handlePair()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/DfsXRkDxVeH2xmme1/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "DfsXRkDxVeH2xmme1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"img1":{"token":"DfsXRkDxVeH2xmme5","src":"/aye-and-nay/albums/DfsXRkDxVeH2xmme1/images/DfsXRkDxVeH2xmme2"},"img2":{"token":"DfsXRkDxVeH2xmme6","src":"/aye-and-nay/albums/DfsXRkDxVeH2xmme1/images/DfsXRkDxVeH2xmme3"}}`+"\n")
	})
	t.Run("Negative", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("WTuVh4YDCdZM4af6", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
		fn := contr.handlePair()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/albums/Tgn6aRNbtx85gz6p1/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "Tgn6aRNbtx85gz6p1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 404)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Album Not Found`+"\n")
	})
}

func TestControllerHandleVote(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "MvdZUxVgPD5p6JTa"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		rand.Shuffle = func(n int, swap func(i int, j int)) {
		}
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("3L8E2zrdQtmJKEwa", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
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
		err := multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		fn = contr.handlePair()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/MvdZUxVgPD5p6JTa1/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "MvdZUxVgPD5p6JTa1"}}
		fn(w, r, ps)
		fn = contr.handleVote()
		w = httptest.NewRecorder()
		json := strings.NewReader(`{"album":{"imgFrom":{"token":"MvdZUxVgPD5p6JTa5"},"imgTo":{"token":"MvdZUxVgPD5p6JTa6"}}}`)
		r = httptest.NewRequest("PATCH", "/api/albums/MvdZUxVgPD5p6JTa1/", json)
		r.Header.Set("Content-Type", "application/json; charset=utf-8")
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "MvdZUxVgPD5p6JTa1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "")
		CheckBody(t, w, ``)
	})
	t.Run("Negative1", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "xtq8FBDgkbk7nZ88"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		rand.Shuffle = func(n int, swap func(i int, j int)) {
		}
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("xGgXp5Pg5nKvGmBY", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
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
		err := multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		fn = contr.handlePair()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/xtq8FBDgkbk7nZ881/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "xtq8FBDgkbk7nZ881"}}
		fn(w, r, ps)
		fn = contr.handleVote()
		w = httptest.NewRecorder()
		json := strings.NewReader(`{"album":{"imgFrom":{"token":"xtq8FBDgkbk7nZ885"},"imgTo":{"token":"xtq8FBDgkbk7nZ886"}}}`)
		r = httptest.NewRequest("PATCH", "/api/albums/22UkVNQPj9nky2gM1/", json)
		r.Header.Set("Content-Type", "application/json; charset=utf-8")
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "22UkVNQPj9nky2gM1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 404)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Token Not Found`+"\n")
	})
	t.Run("Negative2", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "u5u58akruMytGWch"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		rand.Shuffle = func(n int, swap func(i int, j int)) {
		}
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("RkAD9BHx8mTUBYRj", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
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
		err := multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		fn = contr.handlePair()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/u5u58akruMytGWch1/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "u5u58akruMytGWch1"}}
		fn(w, r, ps)
		fn = contr.handleVote()
		w = httptest.NewRecorder()
		json := strings.NewReader(`{"album":{"imgFrom":{"token":"nRqam343KzeNjA9K6"},"imgTo":{"token":"nRqam343KzeNjA9K7"}}}`)
		r = httptest.NewRequest("PATCH", "/api/albums/u5u58akruMytGWch1/", json)
		r.Header.Set("Content-Type", "application/json; charset=utf-8")
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "u5u58akruMytGWch1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 404)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Token Not Found`+"\n")
	})
}

func TestControllerHandleTop(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "bYCppY8q6qjvXjMZ"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		rand.Shuffle = func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("qCzDFPuY53Y34mdS", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		g, ctx := errgroup.WithContext(ctx)
		heartbeat := make(chan struct{})
		serv.StartWorkingPool(ctx, g, heartbeat)
		contr := newController(&serv)
		fn := contr.handleAlbum()
		w := httptest.NewRecorder()
		body := bytes.Buffer{}
		multi := multipart.NewWriter(&body)
		for _, filename := range []string{"alan.jpg", "john.bmp"} {
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
		err := multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		fn = contr.handlePair()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/bYCppY8q6qjvXjMZ1/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "bYCppY8q6qjvXjMZ1"}}
		fn(w, r, ps)
		fn = contr.handleVote()
		w = httptest.NewRecorder()
		json := strings.NewReader(`{"album":{"imgFrom":{"token":"bYCppY8q6qjvXjMZ4"},"imgTo":{"token":"bYCppY8q6qjvXjMZ5"}}}`)
		r = httptest.NewRequest("PATCH", "/api/albums/bYCppY8q6qjvXjMZ1/", json)
		r.Header.Set("Content-Type", "application/json; charset=utf-8")
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "bYCppY8q6qjvXjMZ1"}}
		fn(w, r, ps)
		<-heartbeat
		fn = contr.handlePair()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/bYCppY8q6qjvXjMZ1/", nil)
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "bYCppY8q6qjvXjMZ1"}}
		fn(w, r, ps)
		fn = contr.handleVote()
		w = httptest.NewRecorder()
		json = strings.NewReader(`{"album":{"imgFrom":{"token":"bYCppY8q6qjvXjMZ6"},"imgTo":{"token":"bYCppY8q6qjvXjMZ7"}}}`)
		r = httptest.NewRequest("PATCH", "/api/albums/bYCppY8q6qjvXjMZ1/", json)
		r.Header.Set("Content-Type", "application/json; charset=utf-8")
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "bYCppY8q6qjvXjMZ1"}}
		fn(w, r, ps)
		<-heartbeat
		fn = contr.handleTop()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/bYCppY8q6qjvXjMZ1/top/", nil)
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "bYCppY8q6qjvXjMZ1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"images":[{"src":"/aye-and-nay/albums/bYCppY8q6qjvXjMZ1/images/bYCppY8q6qjvXjMZ2","rating":0.5},{"src":"/aye-and-nay/albums/bYCppY8q6qjvXjMZ1/images/bYCppY8q6qjvXjMZ3","rating":0.5}]}`+"\n")
	})
	t.Run("Negative", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		sched := service.NewScheduler("732qurKQkxYDsG6L", &mem)
		serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
		contr := newController(&serv)
		fn := contr.handleTop()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/albums/54KXhWeFfSK5WXHL1/top/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "54KXhWeFfSK5WXHL1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 404)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Album Not Found`+"\n")
	})
}

func BenchmarkControllerHandleAlbum(b *testing.B) {
	b.ReportAllocs()
	rand.Id = func() func(int) (string, error) {
		id := "MchqrYMkMPmm5z2t"
		i := 0
		return func(length int) (string, error) {
			i++
			return id + strconv.Itoa(i), nil
		}
	}()
	comp := compressor.NewMock()
	stor := storage.NewMock()
	mem := database.NewMem()
	sched := service.NewScheduler("t46W276nEp59gXyr", &mem)
	serv := service.NewService(&comp, &stor, &mem, &mem, &sched)
	contr := newController(&serv)
	fn := contr.handleAlbum()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		w := httptest.NewRecorder()
		body := bytes.Buffer{}
		multi := multipart.NewWriter(&body)
		for j := 0; j < 333; j++ {
			for _, filename := range []string{"alan.jpg", "john.bmp", "dennis.png"} {
				part, err := multi.CreateFormFile("images", filename)
				if err != nil {
					b.Error(err)
				}
				data, err := ioutil.ReadFile("../../testdata/" + filename)
				if err != nil {
					b.Error(err)
				}
				_, err = part.Write(data)
				if err != nil {
					b.Error(err)
				}
			}
		}
		err := multi.Close()
		if err != nil {
			b.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		b.StartTimer()
		fn(w, r, nil)
	}
}
