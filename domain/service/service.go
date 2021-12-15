package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/base64"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	myrand "github.com/zitryss/aye-and-nay/pkg/rand"
)

func New(
	comp domain.Compresser,
	stor domain.Storager,
	pers domain.Databaser,
	temp domain.Cacher,
	qCalc *QueueCalc,
	qComp *QueueComp,
	qDel *QueueDel,
	opts ...options,
) *Service {
	conf := newServiceConfig()
	s := &Service{
		conf:  conf,
		comp:  comp,
		stor:  stor,
		pers:  pers,
		pair:  temp,
		token: temp,
		queue: struct {
			calc *QueueCalc
			comp *QueueComp
			del  *QueueDel
		}{
			qCalc,
			qComp,
			qDel,
		},
		rand: struct {
			id      func() (uint64, error)
			shuffle func(n int, swap func(i int, j int))
		}{
			myrand.Id,
			rand.Shuffle,
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func NewQueueCalc(q domain.Queuer) *QueueCalc {
	return &QueueCalc{newQueue(0x6CF9, q)}
}

type QueueCalc struct {
	*queue
}

func NewQueueComp(q domain.Queuer) *QueueComp {
	return &QueueComp{newQueue(0xDD66, q)}
}

type QueueComp struct {
	*queue
}

func NewQueueDel(q domain.PQueuer) *QueueDel {
	return &QueueDel{newPQueue(0xCDF9, q)}
}

type QueueDel struct {
	*pqueue
}

type options func(*Service)

func WithRandId(fn func() (uint64, error)) options {
	return func(s *Service) {
		s.rand.id = fn
	}
}

func WithRandShuffle(fn func(int, func(int, int))) options {
	return func(s *Service) {
		s.rand.shuffle = fn
	}
}

func WithHeartbeatCalc(ch chan<- interface{}) options {
	return func(s *Service) {
		s.heartbeat.calc = ch
	}
}

func WithHeartbeatComp(ch chan<- interface{}) options {
	return func(s *Service) {
		s.heartbeat.comp = ch
	}
}

func WithHeartbeatDel(ch chan<- interface{}) options {
	return func(s *Service) {
		s.heartbeat.del = ch
	}
}

type Service struct {
	conf  serviceConfig
	comp  domain.Compresser
	stor  domain.Storager
	pers  domain.Databaser
	pair  domain.Stacker
	token domain.Tokener
	queue struct {
		calc *QueueCalc
		comp *QueueComp
		del  *QueueDel
	}
	rand struct {
		id      func() (uint64, error)
		shuffle func(n int, swap func(i, j int))
	}
	heartbeat struct {
		calc chan<- interface{}
		comp chan<- interface{}
		del  chan<- interface{}
	}
}

func (s *Service) Album(ctx context.Context, ff []model.File, dur time.Duration) (uint64, error) {
	album, err := s.rand.id()
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	imgs := make([]model.Image, 0, len(ff))
	for _, f := range ff {
		image, err := s.rand.id()
		if err != nil {
			return 0x0, errors.Wrap(err)
		}
		src, err := s.stor.Put(ctx, album, image, f)
		if err != nil {
			return 0x0, errors.Wrap(err)
		}
		img := model.Image{}
		img.Id = image
		img.Src = src
		imgs = append(imgs, img)
	}
	edgs := map[uint64]map[uint64]int(nil)
	expires := time.Now().Add(dur)
	if dur == 0 {
		expires = time.Time{}
	}
	alb := model.Album{album, imgs, edgs, expires}
	err = s.pers.SaveAlbum(ctx, alb)
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	err = s.queue.comp.add(ctx, album)
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	if dur != 0 {
		err = s.queue.del.add(ctx, album, expires)
		if err != nil {
			return 0x0, errors.Wrap(err)
		}
	}
	return alb.Id, nil
}

func (s *Service) Progress(ctx context.Context, album uint64) (float64, error) {
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

func (s *Service) Pair(ctx context.Context, album uint64) (model.Image, model.Image, error) {
	image1, image2, err := s.pair.Pop(ctx, album)
	if errors.Is(err, domain.ErrPairNotFound) {
		err = s.genPairs(ctx, album)
		if err != nil {
			return model.Image{}, model.Image{}, errors.Wrap(err)
		}
		image1, image2, err = s.pair.Pop(ctx, album)
	}
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	src1 := ""
	src2 := ""
	token1 := image1
	token2 := image2
	if s.conf.tempLinks {
		token1, err = s.rand.id()
		if err != nil {
			return model.Image{}, model.Image{}, errors.Wrap(err)
		}
		err = s.token.Set(ctx, token1, album, image1)
		if err != nil {
			return model.Image{}, model.Image{}, errors.Wrap(err)
		}
		token1B64 := base64.FromUint64(token1)
		src1 = "/api/images/" + token1B64
		token2, err = s.rand.id()
		if err != nil {
			return model.Image{}, model.Image{}, errors.Wrap(err)
		}
		err = s.token.Set(ctx, token2, album, image2)
		if err != nil {
			return model.Image{}, model.Image{}, errors.Wrap(err)
		}
		token2B64 := base64.FromUint64(token2)
		src2 = "/api/images/" + token2B64
	} else {
		src1, err = s.pers.GetImageSrc(ctx, album, image1)
		if err != nil {
			return model.Image{}, model.Image{}, errors.Wrap(err)
		}
		src2, err = s.pers.GetImageSrc(ctx, album, image2)
		if err != nil {
			return model.Image{}, model.Image{}, errors.Wrap(err)
		}
	}
	img1 := model.Image{Id: image1, Src: src1, Token: token1}
	img2 := model.Image{Id: image2, Src: src2, Token: token2}
	return img1, img2, nil
}

func (s *Service) Image(ctx context.Context, token uint64) (model.File, error) {
	album, image, err := s.token.Get(ctx, token)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	f, err := s.stor.Get(ctx, album, image)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	return f, nil
}

func (s *Service) genPairs(ctx context.Context, album uint64) error {
	images, err := s.pers.GetImagesIds(ctx, album)
	if err != nil {
		return errors.Wrap(err)
	}
	s.rand.shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })
	images = append(images, images[0])
	pairs := make([][2]uint64, 0, len(images)-1)
	for i := 0; i < len(images)-1; i++ {
		image1 := images[i]
		image2 := images[i+1]
		pairs = append(pairs, [2]uint64{image1, image2})
	}
	s.rand.shuffle(len(pairs), func(i, j int) { pairs[i], pairs[j] = pairs[j], pairs[i] })
	err = s.pair.Push(ctx, album, pairs)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (s *Service) Vote(ctx context.Context, album uint64, tokenFrom uint64, tokenTo uint64) error {
	imageFrom := tokenFrom
	imageTo := tokenTo
	err := error(nil)
	if s.conf.tempLinks {
		_, imageFrom, err = s.token.Get(ctx, tokenFrom)
		if err != nil {
			return errors.Wrap(err)
		}
		err = s.token.Del(ctx, tokenFrom)
		if err != nil {
			return errors.Wrap(err)
		}
		_, imageTo, err = s.token.Get(ctx, tokenTo)
		if err != nil {
			return errors.Wrap(err)
		}
		err = s.token.Del(ctx, tokenTo)
		if err != nil {
			return errors.Wrap(err)
		}
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

func (s *Service) Top(ctx context.Context, album uint64) ([]model.Image, error) {
	imgs, err := s.pers.GetImagesOrdered(ctx, album)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return imgs, nil
}

func (s *Service) CleanUp(ctx context.Context) error {
	albs, err := s.pers.AlbumsToBeDeleted(ctx)
	if err != nil {
		return errors.Wrap(err)
	}
	for _, alb := range albs {
		err = s.queue.del.add(ctx, alb.Id, alb.Expires)
		if err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}
