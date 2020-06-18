package storage

import (
	"context"
	"sync"
)

func NewMock() mock {
	return mock{
		files: map[string][]byte{},
	}
}

type mock struct {
	sync.Mutex
	files map[string][]byte
}

func (m *mock) Put(_ context.Context, album string, image string, b []byte) (string, error) {
	m.Lock()
	defer m.Unlock()
	filename := "albums/" + album + "/images/" + image
	m.files[filename] = b
	src := "/aye-and-nay/" + filename
	return src, nil
}

func (m *mock) Get(_ context.Context, album string, image string) ([]byte, error) {
	m.Lock()
	defer m.Unlock()
	filename := "albums/" + album + "/images/" + image
	b := m.files[filename]
	return b, nil
}

func (m *mock) Remove(_ context.Context, album string, image string) error {
	m.Lock()
	defer m.Unlock()
	filename := "albums/" + album + "/images/" + image
	delete(m.files, filename)
	return nil
}
