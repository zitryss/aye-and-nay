package ulimit

import (
	"syscall"
)

func init() {
	SetMax = setMax
}

func setMax() error {
	rLimit := syscall.Rlimit{}
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}
	rLimit.Cur = rLimit.Max
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}
	return nil
}
