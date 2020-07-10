package storage

import (
	"context"
)

func NewMock() mock {
	return mock{}
}

type mock struct {
}

func (m *mock) Put(_ context.Context, album string, image string, _ []byte) (string, error) {
	filename := "albums/" + album + "/images/" + image
	src := "/aye-and-nay/" + filename
	return src, nil
}

func (m *mock) Get(_ context.Context, _ string, _ string) ([]byte, error) {
	return nil, nil
}

func (m *mock) Remove(_ context.Context, _ string, _ string) error {
	return nil
}
