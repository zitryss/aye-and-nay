package service

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/internal/log"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func handleError(err error) {
	HandleInnerError(context.Background(), err)
}

func HandleInnerError(ctx context.Context, err error) {
	cause := errors.Cause(err)
	if e := domain.Error(nil); errors.As(cause, &e) {
		log.Print(ctx, e.Inner().Level, "err", err)
		return
	}
	switch cause {
	case context.Canceled:
		log.Debug(ctx, "err", err)
	case context.DeadlineExceeded:
		log.Debug(ctx, "err", err)
	default:
		log.Error(ctx, "err", err)
	}
}
