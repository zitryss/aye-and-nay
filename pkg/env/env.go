package env

import (
	"errors"
	"os"
)

func Lookup(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", errors.New("environment variable " + key + " not found")
	}
	if val == "" {
		return "", errors.New("environment variable " + key + " is empty")
	}
	return val, nil
}
