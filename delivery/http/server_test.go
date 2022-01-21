package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	"github.com/zitryss/aye-and-nay/internal/client"
)

func TestServer(t *testing.T) {
	comp := compressor.NewMock()
	stor := storage.NewMock()
	data := database.NewMem(database.DefaultMemConfig)
	cach := cache.NewMem(cache.DefaultMemConfig)
	qCalc := &service.QueueCalc{}
	qComp := &service.QueueComp{}
	qDel := &service.QueueDel{}
	serv := service.New(service.DefaultServiceConfig, comp, stor, data, cach, qCalc, qComp, qDel)

	middle := NewMiddleware(DefaultMiddlewareConfig, cach)
	srvWait := make(chan error, 1)
	srv, err := NewServer(DefaultServerConfig, middle.Chain, serv, srvWait)
	if err != nil {
		t.Fatal(err)
	}

	mockserver := httptest.NewServer(srv.srv.Handler)
	defer mockserver.Close()
	c, err := client.New(mockserver.URL, 5*time.Second, client.WithFiles("../../testdata"), client.WithTimes(1))
	if err != nil {
		t.Fatal(err)
	}

	album, err := c.Album()
	if err != nil {
		t.Error(err)
	}
	err = c.Ready(album)
	if err != nil {
		t.Error(err)
	}
	p, err := c.Pair(album)
	if err != nil {
		t.Error(err)
	}
	err = c.Do(http.MethodGet, mockserver.URL+p.One.Src, http.NoBody)
	if err != nil {
		t.Error(err)
	}
	err = c.Do(http.MethodGet, mockserver.URL+p.Two.Src, http.NoBody)
	if err != nil {
		t.Error(err)
	}
	err = c.Vote(album, p.One.Token, p.Two.Token)
	if err != nil {
		t.Error(err)
	}
	_, err = c.Top(album)
	if err != nil {
		t.Error(err)
	}
	err = c.Health()
	if err != nil {
		t.Error(err)
	}
}
