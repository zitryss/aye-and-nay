package database

import (
	"context"
	"sort"
	"sync"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func NewMem() *Mem {
	conf := newMemConfig()
	return &Mem{
		conf:       conf,
		syncAlbums: syncAlbums{albums: map[uint64]model.Album{}},
	}
}

type Mem struct {
	conf memConfig
	syncAlbums
}

type syncAlbums struct {
	sync.Mutex
	albums map[uint64]model.Album
}

func (m *Mem) SaveAlbum(_ context.Context, alb model.Album) error {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	_, ok := m.albums[alb.Id]
	if ok {
		return errors.Wrap(model.ErrAlbumAlreadyExists)
	}
	edgs := make(map[uint64]map[uint64]int, len(alb.Images))
	for i := range alb.Images {
		img := &alb.Images[i]
		img.Compressed = m.conf.compressed
		edgs[img.Id] = make(map[uint64]int, len(alb.Images))
	}
	alb.Edges = edgs
	m.albums[alb.Id] = alb
	return nil
}

func (m *Mem) CountImages(_ context.Context, album uint64) (int, error) {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return 0, errors.Wrap(model.ErrAlbumNotFound)
	}
	n := len(alb.Images)
	return n, nil
}

func (m *Mem) CountImagesCompressed(_ context.Context, album uint64) (int, error) {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return 0, errors.Wrap(model.ErrAlbumNotFound)
	}
	n := 0
	for _, img := range alb.Images {
		if img.Compressed {
			n++
		}
	}
	return n, nil
}

func (m *Mem) UpdateCompressionStatus(_ context.Context, album uint64, image uint64) error {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return errors.Wrap(model.ErrAlbumNotFound)
	}
	found := false
	for i := range alb.Images {
		img := &alb.Images[i]
		if img.Id == image {
			img.Compressed = true
			found = true
			break
		}
	}
	if !found {
		return errors.Wrap(model.ErrImageNotFound)
	}
	return nil
}

func (m *Mem) GetImageSrc(_ context.Context, album uint64, image uint64) (string, error) {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return "", errors.Wrap(model.ErrAlbumNotFound)
	}
	found := false
	index := -1
	for i, img := range alb.Images {
		if img.Id == image {
			found = true
			index = i
			break
		}
	}
	if !found {
		return "", errors.Wrap(model.ErrImageNotFound)
	}
	return alb.Images[index].Src, nil
}

func (m *Mem) GetImagesIds(_ context.Context, album uint64) ([]uint64, error) {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return nil, errors.Wrap(model.ErrAlbumNotFound)
	}
	images := make([]uint64, 0, len(alb.Images))
	for _, img := range alb.Images {
		images = append(images, img.Id)
	}
	return images, nil
}

func (m *Mem) SaveVote(_ context.Context, album uint64, imageFrom uint64, imageTo uint64) error {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return errors.Wrap(model.ErrAlbumNotFound)
	}
	alb.Edges[imageFrom][imageTo]++
	return nil
}

func (m *Mem) GetEdges(_ context.Context, album uint64) (map[uint64]map[uint64]int, error) {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return nil, errors.Wrap(model.ErrAlbumNotFound)
	}
	edgs := make(map[uint64]map[uint64]int, len(alb.Edges))
	for k := range alb.Edges {
		edgs[k] = make(map[uint64]int, len(alb.Edges[k]))
	}
	for k1 := range alb.Edges {
		for k2 := range alb.Edges[k1] {
			edgs[k1][k2] = alb.Edges[k1][k2]
		}
	}
	return edgs, nil
}

func (m *Mem) UpdateRatings(_ context.Context, album uint64, vector map[uint64]float64) error {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return errors.Wrap(model.ErrAlbumNotFound)
	}
	for id, rating := range vector {
		for i := range alb.Images {
			img := &alb.Images[i]
			if img.Id == id {
				img.Rating = rating
				break
			}
		}
	}
	return nil
}

func (m *Mem) GetImagesOrdered(_ context.Context, album uint64) ([]model.Image, error) {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return nil, errors.Wrap(model.ErrAlbumNotFound)
	}
	imgs := make([]model.Image, len(alb.Images))
	copy(imgs, alb.Images)
	sort.Slice(imgs, func(i, j int) bool { return imgs[i].Rating > imgs[j].Rating })
	return imgs, nil
}

func (m *Mem) DeleteAlbum(_ context.Context, album uint64) error {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	_, ok := m.albums[album]
	if !ok {
		return errors.Wrap(model.ErrAlbumNotFound)
	}
	delete(m.albums, album)
	return nil
}

func (m *Mem) AlbumsToBeDeleted(_ context.Context) ([]model.Album, error) {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	albs := []model.Album(nil)
	for _, alb := range m.albums {
		if !alb.Expires.IsZero() {
			albs = append(albs, alb)
		}
	}
	return albs, nil
}
