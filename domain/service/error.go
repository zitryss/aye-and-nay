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
	e := errors.Cause(err)
	in := domain.Inner(nil)
	if errors.As(e, &in) {
		in := in.Inner()
		log.Println(log.Level(in.Level), in.DevMsg)
		return
	}
	switch e {
	case context.Canceled:
		log.Debug(e)
	case context.DeadlineExceeded:
		log.Debug(e)
	default:
		log.Errorf("%T %v", e, e)
	}
}
