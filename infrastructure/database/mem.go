package database

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/emirpasic/gods/sets/linkedhashset"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func NewMem(opts ...options) mem {
	conf := newMemConfig()
	m := mem{
		conf:       conf,
		syncAlbums: syncAlbums{albums: map[string]model.Album{}},
		syncQueues: syncQueues{queues: map[string]*linkedhashset.Set{}},
		syncPairs:  syncPairs{pairs: map[string]*pairsTime{}},
		syncTokens: syncTokens{tokens: map[string]*tokenTime{}},
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

type options func(*mem)

func WithHeartbeatPair(ch chan<- interface{}) options {
	return func(m *mem) {
		m.heartbeat.pair = ch
	}
}

func WithHeartbeatToken(ch chan<- interface{}) options {
	return func(m *mem) {
		m.heartbeat.token = ch
	}
}

type mem struct {
	conf memConfig
	syncAlbums
	syncQueues
	syncPairs
	syncTokens
	heartbeat struct {
		pair  chan<- interface{}
		token chan<- interface{}
	}
}

type syncAlbums struct {
	sync.Mutex
	albums map[string]model.Album
}

type syncQueues struct {
	sync.Mutex
	queues map[string]*linkedhashset.Set
}

type syncPairs struct {
	sync.Mutex
	pairs map[string]*pairsTime
}

type pairsTime struct {
	pairs [][2]string
	seen  time.Time
}

type syncTokens struct {
	sync.Mutex
	tokens map[string]*tokenTime
}

type tokenTime struct {
	token string
	seen  time.Time
}

func (m *mem) Monitor() {
	go func() {
		for {
			if m.heartbeat.pair != nil {
				m.heartbeat.pair <- struct{}{}
			}
			now := time.Now()
			m.syncPairs.Lock()
			for k, v := range m.pairs {
				if now.Sub(v.seen) >= m.conf.timeToLive {
					delete(m.pairs, k)
				}
			}
			m.syncPairs.Unlock()
			time.Sleep(m.conf.cleanupInterval)
			if m.heartbeat.pair != nil {
				m.heartbeat.pair <- struct{}{}
			}
		}
	}()
	go func() {
		for {
			if m.heartbeat.token != nil {
				m.heartbeat.token <- struct{}{}
			}
			now := time.Now()
			m.syncTokens.Lock()
			for k, v := range m.tokens {
				if now.Sub(v.seen) >= m.conf.timeToLive {
					delete(m.tokens, k)
				}
			}
			m.syncTokens.Unlock()
			time.Sleep(m.conf.cleanupInterval)
			if m.heartbeat.token != nil {
				m.heartbeat.token <- struct{}{}
			}
		}
	}()
}

func (m *mem) SaveAlbum(_ context.Context, alb model.Album) error {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	_, ok := m.albums[alb.Id]
	if ok {
		return errors.Wrap(model.ErrAlbumAlreadyExists)
	}
	edgs := make(map[string]map[string]int, len(alb.Images))
	for _, img := range alb.Images {
		edgs[img.Id] = make(map[string]int, len(alb.Images))
	}
	alb.Edges = edgs
	m.albums[alb.Id] = alb
	return nil
}

func (m *mem) CountImages(_ context.Context, album string) (int, error) {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return 0, errors.Wrap(model.ErrAlbumNotFound)
	}
	n := len(alb.Images)
	return n, nil
}

func (m *mem) CountImagesCompressed(_ context.Context, album string) (int, error) {
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

func (m *mem) UpdateCompressionStatus(_ context.Context, album string, image string) error {
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

func (m *mem) GetImage(_ context.Context, album string, image string) (model.Image, error) {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return model.Image{}, errors.Wrap(model.ErrAlbumNotFound)
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
		return model.Image{}, errors.Wrap(model.ErrImageNotFound)
	}
	return alb.Images[index], nil
}

func (m *mem) GetImages(_ context.Context, album string) ([]string, error) {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return nil, errors.Wrap(model.ErrAlbumNotFound)
	}
	images := make([]string, 0, len(alb.Images))
	for _, img := range alb.Images {
		images = append(images, img.Id)
	}
	return images, nil
}

func (m *mem) SaveVote(_ context.Context, album string, imageFrom string, imageTo string) error {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return errors.Wrap(model.ErrAlbumNotFound)
	}
	alb.Edges[imageFrom][imageTo]++
	return nil
}

func (m *mem) GetEdges(_ context.Context, album string) (map[string]map[string]int, error) {
	m.syncAlbums.Lock()
	defer m.syncAlbums.Unlock()
	alb, ok := m.albums[album]
	if !ok {
		return nil, errors.Wrap(model.ErrAlbumNotFound)
	}
	edgs := make(map[string]map[string]int, len(alb.Edges))
	for k := range alb.Edges {
		edgs[k] = make(map[string]int, len(alb.Edges[k]))
	}
	for k1 := range alb.Edges {
		for k2 := range alb.Edges[k1] {
			edgs[k1][k2] = alb.Edges[k1][k2]
		}
	}
	return edgs, nil
}

func (m *mem) UpdateRatings(_ context.Context, album string, vector map[string]float64) error {
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

func (m *mem) GetImagesOrdered(_ context.Context, album string) ([]model.Image, error) {
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

func (m *mem) Add(_ context.Context, queue string, album string) error {
	m.syncQueues.Lock()
	defer m.syncQueues.Unlock()
	q, ok := m.queues[queue]
	if !ok {
		q = linkedhashset.New()
		m.queues[queue] = q
	}
	q.Add(album)
	return nil
}

func (m *mem) Poll(_ context.Context, queue string) (string, error) {
	m.syncQueues.Lock()
	defer m.syncQueues.Unlock()
	q, ok := m.queues[queue]
	if !ok {
		return "", errors.Wrap(model.ErrUnknown)
	}
	it := q.Iterator()
	it.Next()
	album := it.Value().(string)
	q.Remove(album)
	return album, nil
}

func (m *mem) Size(_ context.Context, queue string) (int, error) {
	m.syncQueues.Lock()
	defer m.syncQueues.Unlock()
	q, ok := m.queues[queue]
	if !ok {
		return 0, nil
	}
	n := q.Size()
	return n, nil
}

func (m *mem) Push(_ context.Context, album string, pairs [][2]string) error {
	m.syncPairs.Lock()
	defer m.syncPairs.Unlock()
	key := "album:" + album + ":pairs"
	p, ok := m.pairs[key]
	if !ok {
		p = &pairsTime{}
		p.pairs = [][2]string{}
		m.pairs[key] = p
	}
	for _, images := range pairs {
		p.pairs = append(p.pairs, [2]string{images[0], images[1]})
	}
	p.seen = time.Now()
	return nil
}

func (m *mem) Pop(_ context.Context, album string) (string, string, error) {
	m.syncPairs.Lock()
	defer m.syncPairs.Unlock()
	key := "album:" + album + ":pairs"
	p, ok := m.pairs[key]
	if !ok {
		return "", "", errors.Wrap(model.ErrPairNotFound)
	}
	if len(p.pairs) == 0 {
		return "", "", errors.Wrap(model.ErrPairNotFound)
	}
	images := (p.pairs)[0]
	p.pairs = (p.pairs)[1:]
	p.seen = time.Now()
	return images[0], images[1], nil
}

func (m *mem) Set(_ context.Context, album string, token string, image string) error {
	m.syncTokens.Lock()
	defer m.syncTokens.Unlock()
	key := "album:" + album + ":token:" + token + ":image"
	_, ok := m.tokens[key]
	if ok {
		return errors.Wrap(model.ErrTokenAlreadyExists)
	}
	t := &tokenTime{}
	t.token = image
	t.seen = time.Now()
	m.tokens[key] = t
	return nil
}

func (m *mem) Get(_ context.Context, album string, token string) (string, error) {
	m.syncTokens.Lock()
	defer m.syncTokens.Unlock()
	key := "album:" + album + ":token:" + token + ":image"
	image, ok := m.tokens[key]
	if !ok {
		return "", errors.Wrap(model.ErrTokenNotFound)
	}
	delete(m.tokens, key)
	return image.token, nil
}
