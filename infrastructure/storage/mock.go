package storage

import (
	"context"
)

func NewMock() mock {
	return mock{}
}

type mock struct {
}

func (m *mock) Put(_ context.Context, album string, image string, b []byte) (string, error) {
	return "", nil
}

func (m *mock) Get(_ context.Context, album string, image string) ([]byte, error) {
	return nil, nil
}

func (m *mock) Remove(_ context.Context, album string, image string) error {
	return nil
}
