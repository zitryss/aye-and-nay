package service

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/linalg"
	"github.com/zitryss/aye-and-nay/pkg/rand"
)

func NewService(
	comp model.Compresser,
	stor model.Storager,
	pers model.Persister,
	cache model.Cacher,
	sched *scheduler,
) service {
	conf := newServiceConfig()
	serv := service{conf, comp, stor, pers, cache, sched}
	return serv
}

type service struct {
	conf  serviceConfig
	comp  model.Compresser
	stor  model.Storager
	pers  model.Persister
	cache model.Cacher
	sched *scheduler // don't copy sync primitives
}

func (s *service) StartWorkingPool(ctx context.Context, g *errgroup.Group, heartbeat chan<- struct{}) {
	go func() {
		sem := make(chan struct{}, s.conf.numberOfWorkers)
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
					album := s.sched.get()
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
					if heartbeat != nil {
						heartbeat <- struct{}{}
					}
				}
			})
		}
	}()
}

func (s *service) Album(ctx context.Context, files [][]byte) (string, error) {
	id, err := rand.Id(s.conf.albumIdLength)
	if err != nil {
		return "", errors.Wrap(err)
	}
	imgs := make([]model.Image, 0, len(files))
	for _, b := range files {
		id, err := rand.Id(s.conf.imageIdLength)
		if err != nil {
			return "", errors.Wrap(err)
		}
		img := model.Image{}
		img.Id = id
		img.B = b
		imgs = append(imgs, img)
	}
	err = s.comp.Compress(ctx, imgs)
	if errors.Is(err, model.ErrThirdPartyUnavailable) {
		comp := compressor.NewMock()
		s.comp = &comp
	}
	if err != nil {
		return "", errors.Wrap(err)
	}
	err = s.stor.Upload(ctx, id, imgs)
	if err != nil {
		return "", errors.Wrap(err)
	}
	edgs := map[string]map[string]int(nil)
	alb := model.Album{id, imgs, edgs}
	err = s.pers.SaveAlbum(ctx, alb)
	if err != nil {
		return "", errors.Wrap(err)
	}
	return alb.Id, nil
}

func (s *service) Pair(ctx context.Context, album string) (model.Image, model.Image, error) {
	image1, image2, err := s.cache.PopPair(ctx, album)
	if errors.Is(err, model.ErrPairNotFound) {
		err = s.genPairs(ctx, album)
		if err != nil {
			return model.Image{}, model.Image{}, errors.Wrap(err)
		}
		image1, image2, err = s.cache.PopPair(ctx, album)
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
	token1, err := rand.Id(s.conf.tokenIdLength)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	err = s.cache.SetToken(ctx, album, token1, img1.Id)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	img1.Token = token1
	token2, err := rand.Id(s.conf.tokenIdLength)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	err = s.cache.SetToken(ctx, album, token2, img2.Id)
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
	rand.Shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })
	images = append(images, images[0])
	pairs := make([][2]string, 0, len(images)-1)
	for i := 0; i < len(images)-1; i++ {
		image1 := images[i]
		image2 := images[i+1]
		pairs = append(pairs, [2]string{image1, image2})
	}
	rand.Shuffle(len(pairs), func(i, j int) { pairs[i], pairs[j] = pairs[j], pairs[i] })
	err = s.cache.PushPair(ctx, album, pairs)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (s *service) Vote(ctx context.Context, album string, tokenFrom string, tokenTo string) error {
	imageFrom, err := s.cache.GetImageId(ctx, album, tokenFrom)
	if err != nil {
		return errors.Wrap(err)
	}
	imageTo, err := s.cache.GetImageId(ctx, album, tokenTo)
	if err != nil {
		return errors.Wrap(err)
	}
	err = s.pers.SaveVote(ctx, album, imageFrom, imageTo)
	if err != nil {
		return errors.Wrap(err)
	}
	s.sched.put(album)
	return nil
}

func (s *service) Top(ctx context.Context, album string) ([]model.Image, error) {
	imgs, err := s.pers.GetImagesOrdered(ctx, album)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return imgs, nil
}

func (s *service) Exists(ctx context.Context, album string) (bool, error) {
	found, err := s.pers.CheckAlbum(ctx, album)
	if err != nil {
		return false, errors.Wrap(err)
	}
	return found, nil
}

func NewScheduler() scheduler {
	return scheduler{cond: sync.NewCond(&sync.Mutex{})}
}

type scheduler struct {
	cond   *sync.Cond
	closed bool
	ids    []string
}

func (s *scheduler) Monitor(ctx context.Context) {
	go func() {
		<-ctx.Done()
		s.close()
	}()
}

func (s *scheduler) close() {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	s.closed = true
	s.cond.Broadcast()
}

func (s *scheduler) put(newId string) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	for _, id := range s.ids {
		if id == newId {
			return
		}
	}
	s.ids = append(s.ids, newId)
	s.cond.Signal()
}

func (s *scheduler) get() string {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	for len(s.ids) == 0 {
		s.cond.Wait()
		if s.closed {
			return ""
		}
	}
	id := s.ids[0]
	s.ids = s.ids[1:]
	return id
}
