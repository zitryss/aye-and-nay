package errors

import (
	"errors"
	"fmt"
	"runtime"
)

func New(text string) error {
	return errors.New(text)
}

func Wrap(err error) error {
	if err == nil {
		return err
	}
	pc, _, line, ok := runtime.Caller(1)
	if !ok {
		return err
	}
	funcName := runtime.FuncForPC(pc).Name()
	return fmt.Errorf("%s:%d: %w", funcName, line, err)
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return err
	}
	err = fmt.Errorf(format+": %w", append(args, err)...)
	pc, _, line, ok := runtime.Caller(1)
	if !ok {
		return err
	}
	funcName := runtime.FuncForPC(pc).Name()
	return fmt.Errorf("%s:%d: %w", funcName, line, err)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func Cause(err error) error {
	cause := err
	u := cause
	for u != nil {
		cause = u
		u = errors.Unwrap(u)
	}
	return cause
}
