package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func newQueue(id uint64, q domain.Queuer) *queue {
	return &queue{
		id:     id,
		queue:  q,
		cond:   sync.NewCond(&sync.Mutex{}),
		closed: false,
		valid:  true,
	}
}

type queue struct {
	id    uint64
	queue domain.Queuer
	cond  *sync.Cond
	closed bool
	valid  bool
}

func (q *queue) Monitor(ctx context.Context) {
	if q == nil || !q.valid {
		return
	}
	go func() {
		<-ctx.Done()
		q.cond.L.Lock()
		defer q.cond.L.Unlock()
		q.closed = true
		q.cond.Broadcast()
	}()
}

func (q *queue) add(ctx context.Context, album uint64) error {
	if q == nil || !q.valid {
		return nil
	}
	err := q.queue.Add(ctx, q.id, album)
	if err != nil {
		return errors.Wrap(err)
	}
	q.cond.Signal()
	return nil
}

func (q *queue) poll(ctx context.Context) (uint64, error) {
	if q == nil || !q.valid {
		return 0x0, nil
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	n, err := q.queue.Size(ctx, q.id)
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	for n == 0 {
		q.cond.Wait()
		if q.closed {
			return 0x0, nil
		}
		n, err = q.queue.Size(ctx, q.id)
		if err != nil {
			return 0x0, errors.Wrap(err)
		}
	}
	album, err := q.queue.Poll(ctx, q.id)
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	return album, nil
}

func newPQueue(id uint64, pq domain.PQueuer) *pqueue {
	return &pqueue{
		id:      id,
		pqueue:  pq,
		addCh:   make(chan struct{}),
		addBuff: make(chan struct{}, 1),
		done:    0,
		closed:  make(chan struct{}),
		valid:   true,
	}
}

type pqueue struct {
	id     uint64
	pqueue domain.PQueuer
	addCh  chan struct{}
	addBuff chan struct{}
	done    uint32
	closed  chan struct{}
	valid   bool
}

func (pq *pqueue) Monitor(ctx context.Context) {
	if pq == nil || !pq.valid {
		return
	}
	go func() {
		<-ctx.Done()
		pq.closed <- struct{}{}
	}()
	go func() {
		for {
			<-pq.addBuff
			pq.addCh <- struct{}{}
			atomic.StoreUint32(&pq.done, 0)
		}
	}()
}

func (pq *pqueue) add(ctx context.Context, album uint64, expires time.Time) error {
	if pq == nil || !pq.valid {
		return nil
	}
	err := pq.pqueue.PAdd(ctx, pq.id, album, expires)
	if err != nil {
		return errors.Wrap(err)
	}
	if atomic.CompareAndSwapUint32(&pq.done, 0, 1) {
		pq.addBuff <- struct{}{}
	}
	return nil
}

func (pq *pqueue) poll(ctx context.Context) (uint64, error) {
	if pq == nil || !pq.valid {
		return 0x0, nil
	}
	n, err := pq.pqueue.PSize(ctx, pq.id)
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	if n == 0 {
		select {
		case <-pq.closed:
			return 0x0, nil
		case <-pq.addCh:
		}
	}
	select {
	case <-pq.addCh:
	default:
	}
	album, expires, err := pq.pqueue.PPoll(ctx, pq.id)
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	t := time.NewTimer(time.Until(expires))
	defer t.Stop()
	for {
		select {
		case <-pq.closed:
			return 0x0, nil
		case <-pq.addCh:
			newAlbum, newExpires, err := pq.pqueue.PPoll(ctx, pq.id)
			if errors.Is(err, domain.ErrUnknown) {
				err = errors.Wrap(err)
				handleError(err)
				continue
			}
			if err != nil {
				return 0x0, errors.Wrap(err)
			}
			if newExpires.After(expires) {
				err := pq.pqueue.PAdd(ctx, pq.id, newAlbum, newExpires)
				if err != nil {
					return 0x0, errors.Wrap(err)
				}
				continue
			}
			err = pq.pqueue.PAdd(ctx, pq.id, album, expires)
			if err != nil {
				return 0x0, errors.Wrap(err)
			}
			if !t.Stop() {
				<-t.C
			}
			t.Reset(time.Until(newExpires))
			album = newAlbum
			expires = newExpires
		case <-t.C:
			return album, nil
		}
	}
}
