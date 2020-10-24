package service

import (
	"context"
	"sync"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func NewQueue(name string, q model.Queuer) *Queue {
	return &Queue{
		name:   name,
		cond:   sync.NewCond(&sync.Mutex{}),
		closed: false,
		queue:  q,
		valid:  true,
	}
}

type Queue struct {
	name   string
	cond   *sync.Cond
	closed bool
	queue  model.Queuer
	valid  bool
}

func (q *Queue) Monitor(ctx context.Context) {
	go func() {
		<-ctx.Done()
		q.close()
	}()
}

func (q *Queue) close() {
	if q == nil || !q.valid {
		return
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.closed = true
	q.cond.Broadcast()
}

func (q *Queue) add(ctx context.Context, album string) error {
	if q == nil || !q.valid {
		return nil
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	err := q.queue.Add(ctx, q.name, album)
	if err != nil {
		return errors.Wrap(err)
	}
	q.cond.Signal()
	return nil
}

func (q *Queue) poll(ctx context.Context) (string, error) {
	if q == nil || !q.valid {
		return "", nil
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	n, err := q.queue.Size(ctx, q.name)
	if err != nil {
		return "", errors.Wrap(err)
	}
	for n == 0 {
		q.cond.Wait()
		if q.closed {
			return "", nil
		}
		n, err = q.queue.Size(ctx, q.name)
		if err != nil {
			return "", errors.Wrap(err)
		}
	}
	album, err := q.queue.Poll(ctx, q.name)
	if err != nil {
		return "", errors.Wrap(err)
	}
	return album, nil
}
