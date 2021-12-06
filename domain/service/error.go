package service

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func handleError(err error) {
	HandleInnerError(err)
}

func HandleInnerError(err error) {
	cause := errors.Cause(err)
	e := domain.Error(nil)
	if errors.As(cause, &e) {
		log.Println(log.Level(e.Inner().Level), e)
		return
	}
	switch cause {
	case context.Canceled:
		log.Debug(cause)
	case context.DeadlineExceeded:
		log.Debug(cause)
	default:
		log.Errorf("%T %v", cause, cause)
	}
}
