package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func handleHttpRouterError(fn func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		err := fn(w, r, ps)
		if err == nil {
			return
		}
		handleError(w, err)
	}
}

func handleHttpError(fn func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err == nil {
			return
		}
		handleError(w, err)
	}
}

func handleError(w http.ResponseWriter, err error) {
	switch errors.Cause(err) {
	case model.ErrTooManyRequests:
		http.Error(w, "Too Many Requests", 429)
		log.Debug(err)
	case model.ErrForbinden:
		http.Error(w, "Forbidden", 403)
		log.Debug(err)
	case model.ErrPageNotFound:
		http.Error(w, "Page Not Found", 404)
		log.Debug(err)
	case model.ErrNotEnoughImages:
		http.Error(w, "Not Enough Images", 400)
		log.Debug(err)
	case model.ErrTooManyImages:
		http.Error(w, "Request Entity Too Large", 413)
		log.Debug(err)
	case model.ErrImageTooBig:
		http.Error(w, "Request Entity Too Large", 413)
		log.Debug(err)
	case model.ErrNotImage:
		http.Error(w, "Unsupported Media Type", 415)
		log.Debug(err)
	case model.ErrAlbumNotFound:
		http.Error(w, "Album Not Found", 404)
		log.Debug(err)
	case model.ErrTokenNotFound:
		http.Error(w, "Token Not Found", 404)
		log.Debug(err)
	default:
		http.Error(w, "Internal Server Error", 500)
		log.Error(err)
	}
}
