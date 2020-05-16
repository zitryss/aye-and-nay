package service

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/linalg"
	"github.com/zitryss/aye-and-nay/pkg/rand"
)

func NewService(
	comp model.Compresser,
	stor model.Storager,
	pers model.Persister,
	temp model.Temper,
	queue1 *queue,
	queue2 *queue,
) service {
	conf := newServiceConfig()
	return service{
		conf,
		comp,
		stor,
		pers,
		temp,
		temp,
		struct {
			calc *queue
			comp *queue
		}{
			queue1,
			queue2,
		},
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
		calc *queue
		comp *queue
	}
}

func (s *service) StartWorkingPoolCalc(ctx context.Context, g *errgroup.Group, heartbeat chan<- interface{}) {
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
					if heartbeat != nil {
						heartbeat <- struct{}{}
					}
				}
			})
		}
	}()
}

func (s *service) StartWorkingPoolComp(ctx context.Context, g *errgroup.Group, heartbeat chan<- interface{}) {
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
						b, err := s.stor.Get(ctx, album, image)
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue outer
						}
						b, err = s.comp.Compress(ctx, b)
						// if errors.Is(err, model.ErrThirdPartyUnavailable) {
						// 	comp := compressor.NewMock()
						// 	s.comp = &comp
						// }
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
						_, err = s.stor.Put(ctx, album, image, b)
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
						if heartbeat != nil {
							p, _ := s.Progress(context.Background(), album)
							heartbeat <- p
						}
					}
				}
			})
		}
	}()
}

func (s *service) Album(ctx context.Context, files [][]byte) (string, error) {
	album, err := rand.Id(s.conf.albumIdLength)
	if err != nil {
		return "", errors.Wrap(err)
	}
	imgs := make([]model.Image, 0, len(files))
	for _, b := range files {
		image, err := rand.Id(s.conf.imageIdLength)
		if err != nil {
			return "", errors.Wrap(err)
		}
		src, err := s.stor.Put(ctx, album, image, b)
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
	return alb.Id, nil
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
	token1, err := rand.Id(s.conf.tokenIdLength)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	err = s.token.Set(ctx, album, token1, img1.Id)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	img1.Token = token1
	token2, err := rand.Id(s.conf.tokenIdLength)
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
	rand.Shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })
	images = append(images, images[0])
	pairs := make([][2]string, 0, len(images)-1)
	for i := 0; i < len(images)-1; i++ {
		image1 := images[i]
		image2 := images[i+1]
		pairs = append(pairs, [2]string{image1, image2})
	}
	rand.Shuffle(len(pairs), func(i, j int) { pairs[i], pairs[j] = pairs[j], pairs[i] })
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

func (s *service) Exists(ctx context.Context, album string) (bool, error) {
	found, err := s.pers.CheckAlbum(ctx, album)
	if err != nil {
		return false, errors.Wrap(err)
	}
	return found, nil
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

func NewQueue(name string, q model.Queuer) queue {
	return queue{
		name:   name,
		cond:   sync.NewCond(&sync.Mutex{}),
		closed: false,
		queue:  q,
	}
}

type queue struct {
	name   string
	cond   *sync.Cond
	closed bool
	queue  model.Queuer
}

func (q *queue) Monitor(ctx context.Context) {
	go func() {
		<-ctx.Done()
		q.close()
	}()
}

func (q *queue) close() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.closed = true
	q.cond.Broadcast()
}

func (q *queue) add(ctx context.Context, album string) error {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	err := q.queue.Add(ctx, q.name, album)
	if err != nil {
		return errors.Wrap(err)
	}
	q.cond.Signal()
	return nil
}

func (q *queue) poll(ctx context.Context) (string, error) {
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
