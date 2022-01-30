//go:build !windows

package main

import (
	"fmt"
	"os"
	"syscall"
)

func init() {
	err := setUlimit()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "critical:", err)
		os.Exit(1)
	}
}

func setUlimit() error {
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
