package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func newRouter(
	contr controller,
	html html,
) http.Handler {
	router := httprouter.New()
	router.POST("/api/albums/", contr.handleAlbum())
	router.GET("/api/albums/:album/ready", contr.handleReady())
	router.GET("/api/albums/:album/", contr.handlePair())
	router.PATCH("/api/albums/:album/", contr.handleVote())
	router.GET("/api/albums/:album/top/", contr.handleTop())
	router.GET("/", html.handleAlbum())
	router.GET("/albums/:album/", html.handlePair())
	router.GET("/albums/:album/top/", html.handleTop())
	return router
}
