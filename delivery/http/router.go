package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func newRouter(contr controller) http.Handler {
	router := httprouter.New()
	router.POST("/api/albums/", contr.handleAlbum())
	router.GET("/api/albums/:album/ready/", contr.handleReady())
	router.GET("/api/albums/:album/pair/", contr.handlePair())
	router.PATCH("/api/albums/:album/vote/", contr.handleVote())
	router.GET("/api/albums/:album/top/", contr.handleTop())
	return router
}
