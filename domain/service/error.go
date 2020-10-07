package service

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func handleError(err error) {
	switch errors.Cause(err) {
	case model.ErrTooManyRequests:
		log.Debug(err)
	case model.ErrBodyTooLarge:
		log.Debug(err)
	case model.ErrWrongContentType:
		log.Debug(err)
	case model.ErrNotEnoughImages:
		log.Debug(err)
	case model.ErrTooManyImages:
		log.Debug(err)
	case model.ErrImageTooLarge:
		log.Debug(err)
	case model.ErrNotImage:
		log.Debug(err)
	case model.ErrAlbumNotFound:
		log.Debug(err)
	case model.ErrTokenNotFound:
		log.Debug(err)
	case model.ErrThirdPartyUnavailable:
		log.Critical(err)
	case context.Canceled:
		log.Debug(err)
	case context.DeadlineExceeded:
		log.Debug(err)
	default:
		log.Errorf("%T %v", err, err)
	}
}
