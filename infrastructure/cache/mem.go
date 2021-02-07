package cache

import (
	"context"
	"sync"
	"time"

	"github.com/emirpasic/gods/sets/linkedhashset"
	"github.com/emirpasic/gods/trees/binaryheap"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func NewMem(opts ...options) *Mem {
	conf := newMemConfig()
	m := &Mem{
		conf:        conf,
		syncQueues:  syncQueues{queues: map[string]*linkedhashset.Set{}},
		syncPQueues: syncPQueues{pqueues: map[string]*binaryheap.Heap{}},
		syncPairs:   syncPairs{pairs: map[string]*pairsTime{}},
		syncTokens:  syncTokens{tokens: map[string]*tokenTime{}},
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type options func(*Mem)

func WithHeartbeatPair(ch chan<- interface{}) options {
	return func(m *Mem) {
		m.heartbeat.pair = ch
	}
}

func WithHeartbeatToken(ch chan<- interface{}) options {
	return func(m *Mem) {
		m.heartbeat.token = ch
	}
}

type Mem struct {
	conf memConfig
	syncQueues
	syncPQueues
	syncPairs
	syncTokens
	heartbeat struct {
		pair  chan<- interface{}
		token chan<- interface{}
	}
}

type syncQueues struct {
	sync.Mutex
	queues map[string]*linkedhashset.Set
}

type syncPQueues struct {
	sync.Mutex
	pqueues map[string]*binaryheap.Heap
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

type elem struct {
	album   string
	expires time.Time
}

func timeComparator(a, b interface{}) int {
	tA := a.(elem).expires
	tB := b.(elem).expires
	switch {
	case tA.After(tB):
		return 1
	case tA.Before(tB):
		return -1
	default:
		return 0
	}
}

func (m *Mem) Monitor() {
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

func (m *Mem) Add(_ context.Context, queue string, album string) error {
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

func (m *Mem) Poll(_ context.Context, queue string) (string, error) {
	m.syncQueues.Lock()
	defer m.syncQueues.Unlock()
	q, ok := m.queues[queue]
	if !ok {
		return "", errors.Wrap(model.ErrUnknown)
	}
	it := q.Iterator()
	if !it.Next() {
		return "", errors.Wrap(model.ErrUnknown)
	}
	album := it.Value().(string)
	q.Remove(album)
	return album, nil
}

func (m *Mem) Size(_ context.Context, queue string) (int, error) {
	m.syncQueues.Lock()
	defer m.syncQueues.Unlock()
	q, ok := m.queues[queue]
	if !ok {
		return 0, nil
	}
	n := q.Size()
	return n, nil
}

func (m *Mem) PAdd(_ context.Context, pqueue string, album string, expires time.Time) error {
	m.syncPQueues.Lock()
	defer m.syncPQueues.Unlock()
	pq, ok := m.pqueues[pqueue]
	if !ok {
		pq = binaryheap.NewWith(timeComparator)
		m.pqueues[pqueue] = pq
	}
	pq.Push(elem{album, expires})
	return nil
}

func (m *Mem) PPoll(_ context.Context, pqueue string) (string, time.Time, error) {
	m.syncPQueues.Lock()
	defer m.syncPQueues.Unlock()
	pq, ok := m.pqueues[pqueue]
	if !ok {
		return "", time.Time{}, errors.Wrap(model.ErrUnknown)
	}
	e, ok := pq.Pop()
	if !ok {
		return "", time.Time{}, errors.Wrap(model.ErrUnknown)
	}
	return e.(elem).album, e.(elem).expires, nil
}

func (m *Mem) PSize(_ context.Context, pqueue string) (int, error) {
	m.syncPQueues.Lock()
	defer m.syncPQueues.Unlock()
	pq, ok := m.pqueues[pqueue]
	if !ok {
		return 0, nil
	}
	n := pq.Size()
	return n, nil
}

func (m *Mem) Push(_ context.Context, album string, pairs [][2]string) error {
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

func (m *Mem) Pop(_ context.Context, album string) (string, string, error) {
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

func (m *Mem) Set(_ context.Context, album string, token string, image string) error {
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

func (m *Mem) Get(_ context.Context, album string, token string) (string, error) {
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
