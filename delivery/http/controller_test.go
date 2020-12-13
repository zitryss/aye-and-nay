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
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestControllerHandleAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "N2fxX5zbDh8RJQvx"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		ctx := context.Background()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := service.NewQueue("S8Lg9yR7JvfEqQgf", &mem)
		pqueue := (*service.PQueue)(nil)
		heartbeatComp := make(chan interface{})
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, service.WithRandId(fn1), service.WithHeartbeatComp(heartbeatComp))
		g, ctx2 := errgroup.WithContext(ctx)
		serv.StartWorkingPoolComp(ctx2, g)
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
		CheckStatusCode(t, w, 201)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"album":{"id":"N2fxX5zbDh8RJQvx1"}}`+"\n")
		CheckChannel(t, heartbeatComp)
		CheckChannel(t, heartbeatComp)
		CheckChannel(t, heartbeatComp)
		fn = contr.handleReady()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/N2fxX5zbDh8RJQvx1/ready", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "N2fxX5zbDh8RJQvx1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"album":{"progress":1}}`+"\n")
	})
	t.Run("Negative1", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue)
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
		err := multi.WriteField("duration", "1H")
		if err != nil {
			t.Error(err)
		}
		err = multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", "application/json")
		fn(w, r, nil)
		CheckStatusCode(t, w, 415)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Unsupported Content Type"}}`+"\n")
	})
	t.Run("Negative2", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue)
		contr := newController(&serv)
		fn := contr.handleAlbum()
		w := httptest.NewRecorder()
		body := bytes.Buffer{}
		multi := multipart.NewWriter(&body)
		for _, filename := range []string{"linus.jpg", "linus.jpg", "linus.jpg"} {
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
		CheckStatusCode(t, w, 413)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Body Too Large"}}`+"\n")
	})
	t.Run("Negative3", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue)
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
		CheckStatusCode(t, w, 400)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Not Enough Images"}}`+"\n")
	})
	t.Run("Negative4", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue)
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
		CheckStatusCode(t, w, 413)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Too Many Images"}}`+"\n")
	})
	t.Run("Negative5", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue)
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
		CheckStatusCode(t, w, 413)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Image Too Large"}}`+"\n")
	})
	t.Run("Negative6", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue)
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
		CheckStatusCode(t, w, 415)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Unsupported Image Format"}}`+"\n")
	})
	t.Run("Negative7", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "jp8vH6TEapTGgSSc"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		ctx := context.Background()
		heartbeatRestart := make(chan interface{})
		comp := compressor.NewFail(compressor.WithHeartbeatRestart(heartbeatRestart))
		comp.Monitor()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := service.NewQueue("6kD5hhETBcYFbKbq", &mem)
		pqueue := (*service.PQueue)(nil)
		heartbeatComp := make(chan interface{})
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, service.WithRandId(fn1), service.WithHeartbeatComp(heartbeatComp))
		g, ctx2 := errgroup.WithContext(ctx)
		serv.StartWorkingPoolComp(ctx2, g)
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
		CheckStatusCode(t, w, 201)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"album":{"id":"jp8vH6TEapTGgSSc1"}}`+"\n")
		CheckChannel(t, heartbeatComp)
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
		err = multi.WriteField("duration", "1H")
		if err != nil {
			t.Error(err)
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
		CheckBody(t, w, `{"album":{"id":"jp8vH6TEapTGgSSc5"}}`+"\n")
		CheckChannel(t, heartbeatComp)
		CheckChannel(t, heartbeatComp)
		CheckChannel(t, heartbeatComp)
		fn = contr.handleReady()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/jp8vH6TEapTGgSSc1/ready", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "jp8vH6TEapTGgSSc1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"album":{"progress":0}}`+"\n")
		fn = contr.handleReady()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/jp8vH6TEapTGgSSc5/ready", nil)
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "jp8vH6TEapTGgSSc5"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"album":{"progress":1}}`+"\n")
		<-heartbeatRestart
		<-heartbeatRestart
		fn = contr.handleAlbum()
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
		err = multi.WriteField("duration", "1H")
		if err != nil {
			t.Error(err)
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
		CheckBody(t, w, `{"album":{"id":"jp8vH6TEapTGgSSc9"}}`+"\n")
		<-heartbeatComp
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
		err = multi.WriteField("duration", "1H")
		if err != nil {
			t.Error(err)
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
		CheckBody(t, w, `{"album":{"id":"jp8vH6TEapTGgSSc13"}}`+"\n")
		<-heartbeatComp
		<-heartbeatComp
		<-heartbeatComp
		fn = contr.handleReady()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/jp8vH6TEapTGgSSc1/ready", nil)
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "jp8vH6TEapTGgSSc1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"album":{"progress":0}}`+"\n")
		fn = contr.handleReady()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/jp8vH6TEapTGgSSc5/ready", nil)
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "jp8vH6TEapTGgSSc5"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"album":{"progress":1}}`+"\n")
	})
}

func TestControllerHandlePair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "DfsXRkDxVeH2xmme"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, service.WithRandId(fn1), service.WithRandShuffle(fn2))
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
		fn = contr.handlePair()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/DfsXRkDxVeH2xmme1/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "DfsXRkDxVeH2xmme1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"album":{"img1":{"token":"DfsXRkDxVeH2xmme5","src":"/aye-and-nay/albums/DfsXRkDxVeH2xmme1/images/DfsXRkDxVeH2xmme2"},"img2":{"token":"DfsXRkDxVeH2xmme6","src":"/aye-and-nay/albums/DfsXRkDxVeH2xmme1/images/DfsXRkDxVeH2xmme3"}}}`+"\n")
	})
	t.Run("Negative", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue)
		contr := newController(&serv)
		fn := contr.handlePair()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/albums/Tgn6aRNbtx85gz6p1/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "Tgn6aRNbtx85gz6p1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 404)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Album Not Found"}}`+"\n")
	})
}

func TestControllerHandleVote(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "MvdZUxVgPD5p6JTa"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, service.WithRandId(fn1), service.WithRandShuffle(fn2))
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
		fn1 := func() func(int) (string, error) {
			id := "7deCNcaJXzV8jvKP"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, service.WithRandId(fn1), service.WithRandShuffle(fn2))
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
		fn = contr.handlePair()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/7deCNcaJXzV8jvKP1/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "7deCNcaJXzV8jvKP1"}}
		fn(w, r, ps)
		fn = contr.handleVote()
		w = httptest.NewRecorder()
		json := strings.NewReader(`{"album":{"imgFrom":{"token":"7deCNcaJXzV8jvKP5"},"imgTo":{"token":"7deCNcaJXzV8jvKP6"}}}`)
		r = httptest.NewRequest("PATCH", "/api/albums/7deCNcaJXzV8jvKP1/", json)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "7deCNcaJXzV8jvKP1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 415)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Unsupported Content Type"}}`+"\n")
	})
	t.Run("Negative2", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "xtq8FBDgkbk7nZ88"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, service.WithRandId(fn1), service.WithRandShuffle(fn2))
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
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Token Not Found"}}`+"\n")
	})
	t.Run("Negative3", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "u5u58akruMytGWch"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, service.WithRandId(fn1), service.WithRandShuffle(fn2))
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
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Token Not Found"}}`+"\n")
	})
}

func TestControllerHandleTop(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "bYCppY8q6qjvXjMZ"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := service.NewQueue("qCzDFPuY53Y34mdS", &mem)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		heartbeatCalc := make(chan interface{})
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, service.WithRandId(fn1), service.WithRandShuffle(fn2), service.WithHeartbeatCalc(heartbeatCalc))
		g1, ctx1 := errgroup.WithContext(ctx)
		serv.StartWorkingPoolCalc(ctx1, g1)
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
		<-heartbeatCalc
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
		<-heartbeatCalc
		fn = contr.handleTop()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/bYCppY8q6qjvXjMZ1/top/", nil)
		ps = httprouter.Params{httprouter.Param{Key: "album", Value: "bYCppY8q6qjvXjMZ1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 200)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"album":{"images":[{"src":"/aye-and-nay/albums/bYCppY8q6qjvXjMZ1/images/bYCppY8q6qjvXjMZ2","rating":0.5},{"src":"/aye-and-nay/albums/bYCppY8q6qjvXjMZ1/images/bYCppY8q6qjvXjMZ3","rating":0.5}]}}`+"\n")
	})
	t.Run("Negative", func(t *testing.T) {
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*service.Queue)(nil)
		queue2 := (*service.Queue)(nil)
		pqueue := (*service.PQueue)(nil)
		serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue)
		contr := newController(&serv)
		fn := contr.handleTop()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/albums/54KXhWeFfSK5WXHL1/top/", nil)
		ps := httprouter.Params{httprouter.Param{Key: "album", Value: "54KXhWeFfSK5WXHL1"}}
		fn(w, r, ps)
		CheckStatusCode(t, w, 404)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Album Not Found"}}`+"\n")
	})
}

func TestControllerHandleDelete(t *testing.T) {
	fn1 := func() func(int) (string, error) {
		id := "XmL9qT7kJA9uzZTC"
		i := 0
		return func(length int) (string, error) {
			i++
			return id + strconv.Itoa(i), nil
		}
	}()
	fn2 := func() time.Time {
		return time.Now().Add(-1 * time.Hour).Add(100 * time.Millisecond)
	}
	ctx := context.Background()
	comp := compressor.NewMock()
	stor := storage.NewMock()
	mem := database.NewMem()
	queue1 := (*service.Queue)(nil)
	queue2 := (*service.Queue)(nil)
	pqueue := service.NewPQueue("WTgtJN2TemW3vLcT", &mem)
	pqueue.Monitor(ctx)
	heartbeatDel := make(chan interface{})
	serv := service.NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, service.WithRandId(fn1), service.WithRandNow(fn2), service.WithHeartbeatDel(heartbeatDel))
	g1, ctx1 := errgroup.WithContext(ctx)
	serv.StartWorkingPoolDel(ctx1, g1)
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
	select {
	case <-heartbeatDel:
	case <-time.After(120 * time.Millisecond):
		t.Error("<-time.After(120 * time.Millisecond)")
	}
	fn = contr.handleTop()
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/albums/XmL9qT7kJA9uzZTC1/top/", nil)
	ps := httprouter.Params{httprouter.Param{Key: "album", Value: "XmL9qT7kJA9uzZTC1"}}
	fn(w, r, ps)
	CheckStatusCode(t, w, 404)
	CheckContentType(t, w, "application/json; charset=utf-8")
	CheckBody(t, w, `{"error":{"msg":"Album Not Found"}}`+"\n")
}
