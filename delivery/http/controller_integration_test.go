// +build integration

package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"os"
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
	"github.com/zitryss/aye-and-nay/internal/dockertest"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/env"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func TestMain(m *testing.M) {
	_, err := env.Lookup("CONTINUOUS_INTEGRATION")
	if err != nil {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.Lcritical)
		docker := dockertest.New()
		docker.RunMongo()
		docker.RunRedis()
		docker.RunMinio()
		log.SetOutput(ioutil.Discard)
		code := m.Run()
		docker.Purge()
		os.Exit(code)
	}
	code := m.Run()
	os.Exit(code)
}

func TestControllerIntegrationHandleAlbum(t *testing.T) {
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
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("HVyMn8HuDa8rdkyr", &redis)
		queue2 := service.NewQueue("S8Lg9yR7JvfEqQgf", &redis)
		heartbeatComp := make(chan interface{})
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2, service.WithRandId(fn1), service.WithHeartbeatComp(heartbeatComp))
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
		<-heartbeatComp
		<-heartbeatComp
		<-heartbeatComp
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
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("mEdFrvE3549LDFzx", &redis)
		queue2 := service.NewQueue("5qxFFTgPtLVhhQU7", &redis)
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2)
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
		err = multi.Close()
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
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("xQWGJjTtetde2DdB", &redis)
		queue2 := service.NewQueue("g2YS5KE5KeyGU2bG", &redis)
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2)
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
		err = multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		CheckStatusCode(t, w, 413)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Too Many Images`+"\n")
	})
	t.Run("Negative3", func(t *testing.T) {
		comp := compressor.NewMock()
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("eQupRzAY56Qp5E4U", &redis)
		queue2 := service.NewQueue("959UpyN6T8uYaFeB", &redis)
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2)
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
		err = multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		CheckStatusCode(t, w, 413)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Image Too Large`+"\n")
	})
	t.Run("Negative4", func(t *testing.T) {
		comp := compressor.NewMock()
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("Au4DBRQhyEJV99wh", &redis)
		queue2 := service.NewQueue("Zk3KEUJEjDwcsc8u", &redis)
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2)
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
		err = multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		CheckStatusCode(t, w, 415)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Unsupported Image Format`+"\n")
	})
	t.Run("Negative5", func(t *testing.T) {
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
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("Y5gVnAXu4SUg8qK8", &redis)
		queue2 := service.NewQueue("6kD5hhETBcYFbKbq", &redis)
		heartbeatComp := make(chan interface{})
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2, service.WithRandId(fn1), service.WithHeartbeatComp(heartbeatComp))
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
		<-heartbeatComp
		<-heartbeatComp
		<-heartbeatComp
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

func TestControllerIntegrationHandlePair(t *testing.T) {
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
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("93P3AU2V6RMcFND4", &redis)
		queue2 := service.NewQueue("uq4TPwUqmj2MZaCv", &redis)
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2, service.WithRandId(fn1), service.WithRandShuffle(fn2))
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
		CheckBody(t, w, `{"img1":{"token":"DfsXRkDxVeH2xmme5","src":"/aye-and-nay/albums/DfsXRkDxVeH2xmme1/images/DfsXRkDxVeH2xmme2"},"img2":{"token":"DfsXRkDxVeH2xmme6","src":"/aye-and-nay/albums/DfsXRkDxVeH2xmme1/images/DfsXRkDxVeH2xmme3"}}`+"\n")
	})
	t.Run("Negative", func(t *testing.T) {
		comp := compressor.NewMock()
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("WTuVh4YDCdZM4af6", &redis)
		queue2 := service.NewQueue("FNQjKB4hGJC25PJY", &redis)
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2)
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

func TestControllerIntegrationHandleVote(t *testing.T) {
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
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("3L8E2zrdQtmJKEwa", &redis)
		queue2 := service.NewQueue("L4kKdZpZZuTkSDmH", &redis)
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2, service.WithRandId(fn1), service.WithRandShuffle(fn2))
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
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("xGgXp5Pg5nKvGmBY", &redis)
		queue2 := service.NewQueue("6qNjE2tha2Z8s73H", &redis)
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2, service.WithRandId(fn1), service.WithRandShuffle(fn2))
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
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Token Not Found`+"\n")
	})
	t.Run("Negative2", func(t *testing.T) {
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
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("RkAD9BHx8mTUBYRj", &redis)
		queue2 := service.NewQueue("rY4ZJMbTwQGyDqHK", &redis)
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2, service.WithRandId(fn1), service.WithRandShuffle(fn2))
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
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Token Not Found`+"\n")
	})
}

func TestControllerIntegrationHandleTop(t *testing.T) {
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
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("qCzDFPuY53Y34mdS", &redis)
		queue2 := service.NewQueue("YL3b99PHTrMnfX9c", &redis)
		heartbeatCalc := make(chan interface{})
		heartbeatComp := make(chan interface{})
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2, service.WithRandId(fn1), service.WithRandShuffle(fn2), service.WithHeartbeatCalc(heartbeatCalc), service.WithHeartbeatComp(heartbeatComp))
		g1, ctx1 := errgroup.WithContext(ctx)
		serv.StartWorkingPoolCalc(ctx1, g1)
		g2, ctx2 := errgroup.WithContext(ctx)
		serv.StartWorkingPoolComp(ctx2, g2)
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
		err = multi.Close()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest("POST", "/api/albums/", &body)
		r.Header.Set("Content-Type", multi.FormDataContentType())
		fn(w, r, nil)
		<-heartbeatComp
		<-heartbeatComp
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
		CheckBody(t, w, `{"images":[{"src":"/aye-and-nay/albums/bYCppY8q6qjvXjMZ1/images/bYCppY8q6qjvXjMZ2","rating":0.5},{"src":"/aye-and-nay/albums/bYCppY8q6qjvXjMZ1/images/bYCppY8q6qjvXjMZ3","rating":0.5}]}`+"\n")
	})
	t.Run("Negative", func(t *testing.T) {
		comp := compressor.NewMock()
		minio, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		queue1 := service.NewQueue("732qurKQkxYDsG6L", &redis)
		queue2 := service.NewQueue("eJRsgrtZPVc8RE7q", &redis)
		serv := service.NewService(&comp, &minio, &mongo, &redis, &queue1, &queue2)
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
