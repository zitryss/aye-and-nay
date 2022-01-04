package service

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/linalg"
)

func (s *Service) StartWorkingPoolCalc(ctx context.Context, g *errgroup.Group) {
	go func() {
		sem := make(chan struct{}, s.conf.NumberOfWorkersCalc)
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
						e = errors.Wrapf(domain.ErrUnknown, "%v", v)
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
					vect := linalg.PageRank(edgs, s.conf.Accuracy)
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

func (s *Service) StartWorkingPoolComp(ctx context.Context, g *errgroup.Group) {
	go func() {
		sem := make(chan struct{}, s.conf.NumberOfWorkersComp)
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
						e = errors.Wrapf(domain.ErrUnknown, "%v", v)
					}
				}()
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
					images, err := s.pers.GetImagesIds(ctx, album)
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
							continue
						}
						f, err = s.comp.Compress(ctx, f)
						if errors.Is(err, domain.ErrThirdPartyUnavailable) {
							if s.heartbeat.comp != nil {
								s.heartbeat.comp <- err
							}
						}
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue
						}
						_, err = s.stor.Put(ctx, album, image, f)
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue
						}
						err = s.pers.UpdateCompressionStatus(ctx, album, image)
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue
						}
						if s.heartbeat.comp != nil {
							p, _ := s.Progress(ctx, album)
							s.heartbeat.comp <- p
						}
					}
				}
			})
		}
	}()
}

func (s *Service) StartWorkingPoolDel(ctx context.Context, g *errgroup.Group) {
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
				e = errors.Wrapf(domain.ErrUnknown, "%v", v)
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
			images, err := s.pers.GetImagesIds(ctx, album)
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
				s.heartbeat.del <- album
			}
		}
	})
}
