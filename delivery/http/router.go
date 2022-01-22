package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func newRouter(contr controller) http.Handler {
	router := httprouter.New()
	router.POST("/api/albums/", contr.handleAlbum())
	router.GET("/api/albums/:album/status/", contr.handleStatus())
	router.GET("/api/albums/:album/pair/", contr.handlePair())
	router.GET("/api/images/:token/", contr.handleImage())
	router.PATCH("/api/albums/:album/vote/", contr.handleVote())
	router.GET("/api/albums/:album/top/", contr.handleTop())
	router.GET("/api/health/", contr.handleHealth())
	return router
}
