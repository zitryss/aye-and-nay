package service

import (
	"context"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/linalg"
	myrand "github.com/zitryss/aye-and-nay/pkg/rand"
)

func NewService(
	comp model.Compresser,
	stor model.Storager,
	pers model.Persister,
	temp model.Temper,
	queue1 *Queue,
	queue2 *Queue,
	pqueue *PQueue,
	opts ...options,
) service {
	conf := newServiceConfig()
	s := service{
		conf:  conf,
		comp:  comp,
		stor:  stor,
		pers:  pers,
		pair:  temp,
		token: temp,
		queue: struct {
			calc *Queue
			comp *Queue
			del  *PQueue
		}{
			queue1,
			queue2,
			pqueue,
		},
		rand: struct {
			id      func(length int) (string, error)
			shuffle func(n int, swap func(i int, j int))
			now     func() time.Time
		}{
			myrand.Id,
			rand.Shuffle,
			time.Now,
		},
	}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

type options func(*service)

func WithRandId(fn func(int) (string, error)) options {
	return func(s *service) {
		s.rand.id = fn
	}
}

func WithRandShuffle(fn func(int, func(int, int))) options {
	return func(s *service) {
		s.rand.shuffle = fn
	}
}

func WithRandNow(fn func() time.Time) options {
	return func(s *service) {
		s.rand.now = fn
	}
}

func WithHeartbeatCalc(ch chan<- interface{}) options {
	return func(s *service) {
		s.heartbeat.calc = ch
	}
}

func WithHeartbeatComp(ch chan<- interface{}) options {
	return func(s *service) {
		s.heartbeat.comp = ch
	}
}

func WithHeartbeatDel(ch chan<- interface{}) options {
	return func(s *service) {
		s.heartbeat.del = ch
	}
}

type service struct {
	conf  serviceConfig
	comp  model.Compresser
	stor  model.Storager
	pers  model.Persister
	pair  model.Stacker
	token model.Tokener
	queue struct {
		calc *Queue
		comp *Queue
		del  *PQueue
	}
	rand struct {
		id      func(length int) (string, error)
		shuffle func(n int, swap func(i, j int))
		now     func() time.Time
	}
	heartbeat struct {
		calc chan<- interface{}
		comp chan<- interface{}
		del  chan<- interface{}
	}
}

func (s *service) StartWorkingPoolCalc(ctx context.Context, g *errgroup.Group) {
	go func() {
		sem := make(chan struct{}, s.conf.numberOfWorkersCalc)
		for {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			g.Go(func() (e error) {
				defer func() { <-sem }()
				defer func() {
					v := recover()
					if v == nil {
						return
					}
					err, ok := v.(error)
					if ok {
						e = errors.Wrap(err)
					} else {
						e = errors.Wrapf(model.ErrUnknown, "%v", v)
					}
				}()
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}
					album, err := s.queue.calc.poll(ctx)
					if err != nil {
						err = errors.Wrap(err)
						handleError(err)
						e = err
						continue
					}
					select {
					case <-ctx.Done():
						return
					default:
					}
					edgs, err := s.pers.GetEdges(ctx, album)
					if err != nil {
						err = errors.Wrap(err)
						handleError(err)
						e = err
						continue
					}
					vect := linalg.PageRank(edgs)
					err = s.pers.UpdateRatings(ctx, album, vect)
					if err != nil {
						err = errors.Wrap(err)
						handleError(err)
						e = err
						continue
					}
					if s.heartbeat.calc != nil {
						s.heartbeat.calc <- struct{}{}
					}
				}
			})
		}
	}()
}

func (s *service) StartWorkingPoolComp(ctx context.Context, g *errgroup.Group) {
	go func() {
		sem := make(chan struct{}, s.conf.numberOfWorkersComp)
		for {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			g.Go(func() (e error) {
				defer func() { <-sem }()
				defer func() {
					v := recover()
					if v == nil {
						return
					}
					err, ok := v.(error)
					if ok {
						e = errors.Wrap(err)
					} else {
						e = errors.Wrapf(model.ErrUnknown, "%v", v)
					}
				}()
			outer:
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}
					album, err := s.queue.comp.poll(ctx)
					if err != nil {
						err = errors.Wrap(err)
						handleError(err)
						e = err
						continue
					}
					select {
					case <-ctx.Done():
						return
					default:
					}
					images, err := s.pers.GetImages(ctx, album)
					if err != nil {
						err = errors.Wrap(err)
						handleError(err)
						e = err
						continue
					}
					for _, image := range images {
						f, err := s.stor.Get(ctx, album, image)
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue outer
						}
						f, err = s.comp.Compress(ctx, f)
						if errors.Is(err, model.ErrThirdPartyUnavailable) {
							if s.heartbeat.comp != nil {
								s.heartbeat.comp <- err
							}
						}
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue outer
						}
						err = s.stor.Remove(ctx, album, image)
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue outer
						}
						_, err = s.stor.Put(ctx, album, image, f)
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue outer
						}
						err = s.pers.UpdateCompressionStatus(ctx, album, image)
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue outer
						}
						if s.heartbeat.comp != nil {
							p, _ := s.Progress(context.Background(), album)
							s.heartbeat.comp <- p
						}
					}
				}
			})
		}
	}()
}

func (s *service) StartWorkingPoolDel(ctx context.Context, g *errgroup.Group) {
	g.Go(func() (e error) {
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			err, ok := v.(error)
			if ok {
				e = errors.Wrap(err)
			} else {
				e = errors.Wrapf(model.ErrUnknown, "%v", v)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			album, err := s.queue.del.poll(ctx)
			if err != nil {
				err = errors.Wrap(err)
				handleError(err)
				e = err
				continue
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
			images, err := s.pers.GetImages(ctx, album)
			if err != nil {
				err = errors.Wrap(err)
				handleError(err)
				e = err
				continue
			}
			err = s.pers.DeleteAlbum(ctx, album)
			if err != nil {
				err = errors.Wrap(err)
				handleError(err)
				e = err
				continue
			}
			for _, image := range images {
				err = s.stor.Remove(ctx, album, image)
				if err != nil {
					err = errors.Wrap(err)
					handleError(err)
					e = err
					continue
				}
			}
			if s.heartbeat.del != nil {
				s.heartbeat.del <- struct{}{}
			}
		}
	})
}

func (s *service) Album(ctx context.Context, ff []model.File, dur time.Duration) (string, error) {
	album, err := s.rand.id(s.conf.albumIdLength)
	if err != nil {
		return "", errors.Wrap(err)
	}
	imgs := make([]model.Image, 0, len(ff))
	for _, f := range ff {
		image, err := s.rand.id(s.conf.imageIdLength)
		if err != nil {
			return "", errors.Wrap(err)
		}
		src, err := s.stor.Put(ctx, album, image, f)
		if err != nil {
			return "", errors.Wrap(err)
		}
		img := model.Image{}
		img.Id = image
		img.Src = src
		imgs = append(imgs, img)
	}
	edgs := map[string]map[string]int(nil)
	alb := model.Album{album, imgs, edgs}
	err = s.pers.SaveAlbum(ctx, alb)
	if err != nil {
		return "", errors.Wrap(err)
	}
	err = s.queue.comp.add(ctx, album)
	if err != nil {
		return "", errors.Wrap(err)
	}
	err = s.queue.del.add(ctx, album, s.rand.now().Add(dur))
	if err != nil {
		return "", errors.Wrap(err)
	}
	return alb.Id, nil
}

func (s *service) Progress(ctx context.Context, album string) (float64, error) {
	all, err := s.pers.CountImages(ctx, album)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	n, err := s.pers.CountImagesCompressed(ctx, album)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return float64(n) / float64(all), nil
}

func (s *service) Pair(ctx context.Context, album string) (model.Image, model.Image, error) {
	image1, image2, err := s.pair.Pop(ctx, album)
	if errors.Is(err, model.ErrPairNotFound) {
		err = s.genPairs(ctx, album)
		if err != nil {
			return model.Image{}, model.Image{}, errors.Wrap(err)
		}
		image1, image2, err = s.pair.Pop(ctx, album)
	}
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	img1, err := s.pers.GetImage(ctx, album, image1)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	img2, err := s.pers.GetImage(ctx, album, image2)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	token1, err := s.rand.id(s.conf.tokenIdLength)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	err = s.token.Set(ctx, album, token1, img1.Id)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	img1.Token = token1
	token2, err := s.rand.id(s.conf.tokenIdLength)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	err = s.token.Set(ctx, album, token2, img2.Id)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	img2.Token = token2
	return img1, img2, nil
}

func (s *service) genPairs(ctx context.Context, album string) error {
	images, err := s.pers.GetImages(ctx, album)
	if err != nil {
		return errors.Wrap(err)
	}
	s.rand.shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })
	images = append(images, images[0])
	pairs := make([][2]string, 0, len(images)-1)
	for i := 0; i < len(images)-1; i++ {
		image1 := images[i]
		image2 := images[i+1]
		pairs = append(pairs, [2]string{image1, image2})
	}
	s.rand.shuffle(len(pairs), func(i, j int) { pairs[i], pairs[j] = pairs[j], pairs[i] })
	err = s.pair.Push(ctx, album, pairs)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (s *service) Vote(ctx context.Context, album string, tokenFrom string, tokenTo string) error {
	imageFrom, err := s.token.Get(ctx, album, tokenFrom)
	if err != nil {
		return errors.Wrap(err)
	}
	imageTo, err := s.token.Get(ctx, album, tokenTo)
	if err != nil {
		return errors.Wrap(err)
	}
	err = s.pers.SaveVote(ctx, album, imageFrom, imageTo)
	if err != nil {
		return errors.Wrap(err)
	}
	err = s.queue.calc.add(ctx, album)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (s *service) Top(ctx context.Context, album string) ([]model.Image, error) {
	imgs, err := s.pers.GetImagesOrdered(ctx, album)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return imgs, nil
}
