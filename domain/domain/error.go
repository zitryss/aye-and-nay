package domain

import (
	"fmt"
	"net/http"
)

const (
	ldisabled int = iota
	ldebug
	linfo
	lerror
	lcritical
)

var (
	ErrTooManyRequests = &domainError{
		outerError: outerError{
			StatusCode: http.StatusTooManyRequests,
			AppCode:    0x0,
			UserMsg:    "too many requests",
		},
		innerError: innerError{
			Level:  ldebug,
			DevMsg: "too many requests",
		},
	}
	ErrBodyTooLarge = &domainError{
		outerError: outerError{
			StatusCode: http.StatusRequestEntityTooLarge,
			AppCode:    0x1,
			UserMsg:    "body too large",
		},
		innerError: innerError{
			Level:  ldebug,
			DevMsg: "body too large",
		},
	}
	ErrWrongContentType = &domainError{
		outerError: outerError{
			StatusCode: http.StatusUnsupportedMediaType,
			AppCode:    0x2,
			UserMsg:    "unsupported media type",
		},
		innerError: innerError{
			Level:  ldebug,
			DevMsg: "wrong content type",
		},
	}
	ErrNotEnoughImages = &domainError{
		outerError: outerError{
			StatusCode: http.StatusBadRequest,
			AppCode:    0x3,
			UserMsg:    "not enough images",
		},
		innerError: innerError{
			Level:  ldebug,
			DevMsg: "not enough images",
		},
	}
	ErrTooManyImages = &domainError{
		outerError: outerError{
			StatusCode: http.StatusRequestEntityTooLarge,
			AppCode:    0x4,
			UserMsg:    "too many images",
		},
		innerError: innerError{
			Level:  ldebug,
			DevMsg: "too many images",
		},
	}
	ErrImageTooLarge = &domainError{
		outerError: outerError{
			StatusCode: http.StatusRequestEntityTooLarge,
			AppCode:    0x5,
			UserMsg:    "image too large",
		},
		innerError: innerError{
			Level:  ldebug,
			DevMsg: "image too large",
		},
	}
	ErrNotImage = &domainError{
		outerError: outerError{
			StatusCode: http.StatusUnsupportedMediaType,
			AppCode:    0x6,
			UserMsg:    "unsupported media type",
		},
		innerError: innerError{
			Level:  ldebug,
			DevMsg: "not an image",
		},
	}
	ErrDurationNotSet = &domainError{
		outerError: outerError{
			StatusCode: http.StatusBadRequest,
			AppCode:    0x7,
			UserMsg:    "duration not set",
		},
		innerError: innerError{
			Level:  ldebug,
			DevMsg: "duration not set",
		},
	}
	ErrDurationInvalid = &domainError{
		outerError: outerError{
			StatusCode: http.StatusBadRequest,
			AppCode:    0x8,
			UserMsg:    "duration invalid",
		},
		innerError: innerError{
			Level:  ldebug,
			DevMsg: "duration invalid",
		},
	}
	ErrAlbumNotFound = &domainError{
		outerError: outerError{
			StatusCode: http.StatusNotFound,
			AppCode:    0x9,
			UserMsg:    "album not found",
		},
		innerError: innerError{
			Level:  ldebug,
			DevMsg: "album not found",
		},
	}
	ErrPairNotFound = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0xA,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  lerror,
			DevMsg: "pair not found",
		},
	}
	ErrTokenNotFound = &domainError{
		outerError: outerError{
			StatusCode: http.StatusNotFound,
			AppCode:    0xB,
			UserMsg:    "token not found",
		},
		innerError: innerError{
			Level:  ldebug,
			DevMsg: "token not found",
		},
	}
	ErrImageNotFound = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0xC,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  lerror,
			DevMsg: "image not found",
		},
	}
	ErrAlbumAlreadyExists = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0xD,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  lerror,
			DevMsg: "album already exists",
		},
	}
	ErrTokenAlreadyExists = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0xE,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  lerror,
			DevMsg: "token already exists",
		},
	}
	ErrUnsupportedMediaType = &domainError{
		outerError: outerError{
			StatusCode: http.StatusUnsupportedMediaType,
			AppCode:    0xF,
			UserMsg:    "unsupported media type",
		},
		innerError: innerError{
			Level:  lerror,
			DevMsg: "image rejected by third party",
		},
	}
	ErrThirdPartyUnavailable = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x10,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  lcritical,
			DevMsg: "third party is unavailable",
		},
	}
	ErrBadHealthCompressor = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x11,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  lcritical,
			DevMsg: "compressor is unavailable",
		},
	}
	ErrBadHealthStorage = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x12,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  lcritical,
			DevMsg: "storage is unavailable",
		},
	}
	ErrBadHealthDatabase = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x13,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  lcritical,
			DevMsg: "database is unavailable",
		},
	}
	ErrBadHealthCache = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x14,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  lcritical,
			DevMsg: "cache is unavailable",
		},
	}
	ErrUnknown = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x11,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  lerror,
			DevMsg: "unknown",
		},
	}
)

type Error interface {
	error
	Outer
	Inner
}

type Outer interface {
	Outer() outerError
}

type Inner interface {
	Inner() innerError
}

type domainError struct {
	outerError
	innerError
}

type outerError struct {
	StatusCode int
	AppCode    int
	UserMsg    string
}

type innerError struct {
	Level  int
	DevMsg string
}

func (de *domainError) Error() string {
	return fmt.Sprintf("{%+v, %+v}", de.outerError, de.innerError)
}

func (de *domainError) Outer() outerError {
	return de.outerError
}

func (de *domainError) Inner() innerError {
	return de.innerError
}

func Wrap(err error, sentinel *domainError) error {
	return &wrappedError{original: err, sentinel: sentinel}
}

type wrappedError struct {
	original error
	sentinel *domainError
}

func (we *wrappedError) Is(err error) bool {
	return we.sentinel == err
}

func (we *wrappedError) Error() string {
	return we.original.Error() + ": " + we.sentinel.Error()
}

func (we *wrappedError) Unwrap() error {
	return nil
}

func (we *wrappedError) Outer() outerError {
	return we.sentinel.Outer()
}

func (we *wrappedError) Inner() innerError {
	return we.sentinel.Inner()
}
