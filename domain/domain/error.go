package domain

import (
	"fmt"
	"net/http"
)

const (
	logDisabled int = iota
	logDebug
	logInfo
	logError
	logCritical
)

var (
	ErrTooManyRequests = &domainError{
		outerError: outerError{
			StatusCode: http.StatusTooManyRequests,
			AppCode:    0x1,
			UserMsg:    "too many requests",
		},
		innerError: innerError{
			Level:  logDebug,
			DevMsg: "too many requests",
		},
	}
	ErrBodyTooLarge = &domainError{
		outerError: outerError{
			StatusCode: http.StatusRequestEntityTooLarge,
			AppCode:    0x2,
			UserMsg:    "body too large",
		},
		innerError: innerError{
			Level:  logDebug,
			DevMsg: "body too large",
		},
	}
	ErrWrongContentType = &domainError{
		outerError: outerError{
			StatusCode: http.StatusUnsupportedMediaType,
			AppCode:    0x3,
			UserMsg:    "unsupported media type",
		},
		innerError: innerError{
			Level:  logDebug,
			DevMsg: "wrong content type",
		},
	}
	ErrNotEnoughImages = &domainError{
		outerError: outerError{
			StatusCode: http.StatusBadRequest,
			AppCode:    0x4,
			UserMsg:    "not enough images",
		},
		innerError: innerError{
			Level:  logDebug,
			DevMsg: "not enough images",
		},
	}
	ErrTooManyImages = &domainError{
		outerError: outerError{
			StatusCode: http.StatusRequestEntityTooLarge,
			AppCode:    0x5,
			UserMsg:    "too many images",
		},
		innerError: innerError{
			Level:  logDebug,
			DevMsg: "too many images",
		},
	}
	ErrImageTooLarge = &domainError{
		outerError: outerError{
			StatusCode: http.StatusRequestEntityTooLarge,
			AppCode:    0x6,
			UserMsg:    "image too large",
		},
		innerError: innerError{
			Level:  logDebug,
			DevMsg: "image too large",
		},
	}
	ErrNotImage = &domainError{
		outerError: outerError{
			StatusCode: http.StatusUnsupportedMediaType,
			AppCode:    0x7,
			UserMsg:    "unsupported media type",
		},
		innerError: innerError{
			Level:  logDebug,
			DevMsg: "not an image",
		},
	}
	ErrDurationNotSet = &domainError{
		outerError: outerError{
			StatusCode: http.StatusBadRequest,
			AppCode:    0x8,
			UserMsg:    "duration not set",
		},
		innerError: innerError{
			Level:  logDebug,
			DevMsg: "duration not set",
		},
	}
	ErrDurationInvalid = &domainError{
		outerError: outerError{
			StatusCode: http.StatusBadRequest,
			AppCode:    0x9,
			UserMsg:    "duration invalid",
		},
		innerError: innerError{
			Level:  logDebug,
			DevMsg: "duration invalid",
		},
	}
	ErrAlbumNotFound = &domainError{
		outerError: outerError{
			StatusCode: http.StatusNotFound,
			AppCode:    0xA,
			UserMsg:    "album not found",
		},
		innerError: innerError{
			Level:  logDebug,
			DevMsg: "album not found",
		},
	}
	ErrPairNotFound = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0xB,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  logError,
			DevMsg: "pair not found",
		},
	}
	ErrTokenNotFound = &domainError{
		outerError: outerError{
			StatusCode: http.StatusNotFound,
			AppCode:    0xC,
			UserMsg:    "token not found",
		},
		innerError: innerError{
			Level:  logDebug,
			DevMsg: "token not found",
		},
	}
	ErrImageNotFound = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0xD,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  logError,
			DevMsg: "image not found",
		},
	}
	ErrAlbumAlreadyExists = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0xE,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  logError,
			DevMsg: "album already exists",
		},
	}
	ErrTokenAlreadyExists = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0xF,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  logError,
			DevMsg: "token already exists",
		},
	}
	ErrUnsupportedMediaType = &domainError{
		outerError: outerError{
			StatusCode: http.StatusUnsupportedMediaType,
			AppCode:    0x10,
			UserMsg:    "unsupported media type",
		},
		innerError: innerError{
			Level:  logError,
			DevMsg: "image rejected by third party",
		},
	}
	ErrThirdPartyUnavailable = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x11,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  logCritical,
			DevMsg: "third party is unavailable",
		},
	}
	ErrBadHealthCompressor = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x12,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  logCritical,
			DevMsg: "compressor is unavailable",
		},
	}
	ErrBadHealthStorage = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x13,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  logCritical,
			DevMsg: "storage is unavailable",
		},
	}
	ErrBadHealthDatabase = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x14,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  logCritical,
			DevMsg: "database is unavailable",
		},
	}
	ErrBadHealthCache = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x15,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  logCritical,
			DevMsg: "cache is unavailable",
		},
	}
	ErrUnknown = &domainError{
		outerError: outerError{
			StatusCode: http.StatusInternalServerError,
			AppCode:    0x16,
			UserMsg:    "internal server error",
		},
		innerError: innerError{
			Level:  logError,
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
