package model

import (
	"errors"
)

var (
	ErrTooManyRequests       = errors.New("too many requests")
	ErrNotEnoughImages       = errors.New("not enough images")
	ErrTooManyImages         = errors.New("too many images")
	ErrImageTooBig           = errors.New("image too big")
	ErrNotImage              = errors.New("not image")
	ErrAlbumNotFound         = errors.New("album not found")
	ErrPairNotFound          = errors.New("pair not found")
	ErrTokenNotFound         = errors.New("token not found")
	ErrImageNotFound         = errors.New("image not found")
	ErrAlbumAlreadyExists    = errors.New("album already exists")
	ErrTokenAlreadyExists    = errors.New("token already exists")
	ErrThirdPartyUnavailable = errors.New("third party unavailable")
	ErrUnknown               = errors.New("unknown")
)
