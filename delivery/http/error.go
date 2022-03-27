package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/pkg/errors"
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
	service.HandleInnerError(err)
	handleOuterError(w, err)
}

func handleOuterError(w http.ResponseWriter, err error) {
	resp := errorResponse{}
	defer func() {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(resp.Error.statusCode)
		_ = json.NewEncoder(w).Encode(resp)
	}()
	cause := errors.Cause(err)
	if e := domain.Error(nil); errors.As(cause, &e) {
		out := e.Outer()
		resp.Error.statusCode = out.StatusCode
		resp.Error.AppCode = out.AppCode
		resp.Error.UserMsg = out.UserMsg
		return
	}
	switch cause {
	case context.Canceled:
		resp.Error.statusCode = http.StatusInternalServerError
		resp.Error.AppCode = -1
		resp.Error.UserMsg = "internal server error"
	case context.DeadlineExceeded:
		resp.Error.statusCode = http.StatusInternalServerError
		resp.Error.AppCode = -2
		resp.Error.UserMsg = "internal server error"
	default:
		resp.Error.statusCode = http.StatusInternalServerError
		resp.Error.AppCode = -3
		resp.Error.UserMsg = "internal server error"
	}
}
