//go:build unit

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)

	mockserver := httptest.NewServer(srv.srv.Handler)
	defer mockserver.Close()
	c, err := client.New(mockserver.URL, 5*time.Second, client.WithFiles("../../testdata"), client.WithTimes(1))
	require.NoError(t, err)

	album, err := c.Album()
	assert.NoError(t, err)
	err = c.Status(album)
	assert.NoError(t, err)
	p, err := c.Pair(album)
	assert.NoError(t, err)
	if service.DefaultServiceConfig.TempLinks == true {
		err = c.Do(http.MethodGet, mockserver.URL+p.One.Src, http.NoBody)
		assert.NoError(t, err)
		err = c.Do(http.MethodGet, mockserver.URL+p.Two.Src, http.NoBody)
		assert.NoError(t, err)
	}
	err = c.Vote(album, p.One.Token, p.Two.Token)
	assert.NoError(t, err)
	_, err = c.Top(album)
	assert.NoError(t, err)
	err = c.Health()
	assert.NoError(t, err)
}
