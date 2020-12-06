package http

import (
	"context"
	"encoding/json"
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
	resp := errorResponse{}
	switch errors.Cause(err) {
	case model.ErrTooManyRequests:
		log.Debug(err)
		resp.Error.code = 429
		resp.Error.Msg = "Too Many Requests"
	case model.ErrBodyTooLarge:
		log.Debug(err)
		resp.Error.code = 413
		resp.Error.Msg = "Body Too Large"
	case model.ErrWrongContentType:
		log.Debug(err)
		resp.Error.code = 415
		resp.Error.Msg = "Unsupported Content Type"
	case model.ErrNotEnoughImages:
		log.Debug(err)
		resp.Error.code = 400
		resp.Error.Msg = "Not Enough Images"
	case model.ErrTooManyImages:
		log.Debug(err)
		resp.Error.code = 413
		resp.Error.Msg = "Too Many Images"
	case model.ErrImageTooLarge:
		log.Debug(err)
		resp.Error.code = 413
		resp.Error.Msg = "Image Too Large"
	case model.ErrNotImage:
		log.Debug(err)
		resp.Error.code = 415
		resp.Error.Msg = "Unsupported Image Format"
	case model.ErrDurationNotSet:
		log.Debug(err)
		resp.Error.code = 400
		resp.Error.Msg = "Duration Not Set"
	case model.ErrDurationInvalid:
		log.Debug(err)
		resp.Error.code = 400
		resp.Error.Msg = "Duration Invalid"
	case model.ErrAlbumNotFound:
		log.Debug(err)
		resp.Error.code = 404
		resp.Error.Msg = "Album Not Found"
	case model.ErrTokenNotFound:
		log.Debug(err)
		resp.Error.code = 404
		resp.Error.Msg = "Token Not Found"
	case model.ErrThirdPartyUnavailable:
		log.Critical(err)
		resp.Error.code = 500
		resp.Error.Msg = "Internal Server Error"
	case context.Canceled:
		log.Debug(err)
		resp.Error.code = 500
		resp.Error.Msg = "Internal Server Error"
	case context.DeadlineExceeded:
		log.Debug(err)
		resp.Error.code = 500
		resp.Error.Msg = "Internal Server Error"
	default:
		log.Errorf("%T %v", err, err)
		resp.Error.code = 500
		resp.Error.Msg = "Internal Server Error"
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(resp.Error.code)
	_ = json.NewEncoder(w).Encode(resp)
}
