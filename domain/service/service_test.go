//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestServiceAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0x463E + i, nil
			}
		}()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{newQueue(0xB273, mCache)}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		heartbeatComp := make(chan interface{})
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithHeartbeatComp(heartbeatComp))
		gComp, ctxComp := errgroup.WithContext(ctx)
		serv.StartWorkingPoolComp(ctxComp, gComp)
		files := []model.File{Png(), Png()}
		_, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		v := AssertChannel(t, heartbeatComp)
		p, ok := v.(float64)
		assert.True(t, ok)
		assert.InDelta(t, 0.5, p, TOLERANCE)
		v = AssertChannel(t, heartbeatComp)
		p, ok = v.(float64)
		assert.True(t, ok)
		assert.InDelta(t, 1, p, TOLERANCE)
	})
	t.Run("Negative", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0x915C + i, nil
			}
		}()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		heartbeatRestart := make(chan interface{})
		comp := compressor.NewShortpixel(compressor.DefaultShortpixelConfig, compressor.WithHeartbeatRestart(heartbeatRestart))
		comp.Monitor()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{newQueue(0x88AB, mCache)}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		heartbeatComp := make(chan interface{})
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithHeartbeatComp(heartbeatComp))
		gComp, ctxComp := errgroup.WithContext(ctx)
		serv.StartWorkingPoolComp(ctxComp, gComp)
		files := []model.File{Png(), Png()}
		_, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		v := AssertChannel(t, heartbeatComp)
		_ = AssertChannel(t, heartbeatComp)
		err, ok := v.(error)
		assert.True(t, ok)
		assert.ErrorIs(t, err, domain.ErrThirdPartyUnavailable)
		files = []model.File{Png(), Png()}
		_, err = serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		v = AssertChannel(t, heartbeatComp)
		p, ok := v.(float64)
		assert.True(t, ok)
		assert.InDelta(t, 0.5, p, TOLERANCE)
		v = AssertChannel(t, heartbeatComp)
		p, ok = v.(float64)
		assert.True(t, ok)
		assert.InDelta(t, 1, p, TOLERANCE)
		AssertChannel(t, heartbeatRestart)
		AssertChannel(t, heartbeatRestart)
		files = []model.File{Png(), Png()}
		_, err = serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		v = AssertChannel(t, heartbeatComp)
		_ = AssertChannel(t, heartbeatComp)
		err, ok = v.(error)
		assert.True(t, ok)
		assert.ErrorIs(t, err, domain.ErrThirdPartyUnavailable)
		files = []model.File{Png(), Png()}
		_, err = serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		v = AssertChannel(t, heartbeatComp)
		p, ok = v.(float64)
		assert.True(t, ok)
		assert.InDelta(t, 0.5, p, TOLERANCE)
		v = AssertChannel(t, heartbeatComp)
		p, ok = v.(float64)
		assert.True(t, ok)
		assert.InDelta(t, 1, p, TOLERANCE)
	})
}

func TestServicePair(t *testing.T) {
	t.Run("Positive1", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0x3BC5 + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img7, img8, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img1 := model.Image{Id: 0x3BC7, Token: 0x3BC9, Src: "/api/images/yTsAAAAAAAA/"}
		img2 := model.Image{Id: 0x3BC8, Token: 0x3BCA, Src: "/api/images/yjsAAAAAAAA/"}
		imgs1 := []model.Image{img1, img2}
		assert.NotEqual(t, img7, img8)
		AssertContains(t, imgs1, img7)
		AssertContains(t, imgs1, img8)
		img9, img10, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img3 := model.Image{Id: 0x3BC8, Token: 0x3BCB, Src: "/api/images/yzsAAAAAAAA/"}
		img4 := model.Image{Id: 0x3BC7, Token: 0x3BCC, Src: "/api/images/zDsAAAAAAAA/"}
		imgs2 := []model.Image{img3, img4}
		assert.NotEqual(t, img9, img10)
		AssertContains(t, imgs2, img9)
		AssertContains(t, imgs2, img10)
		img11, img12, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img5 := model.Image{Id: 0x3BC7, Token: 0x3BCD, Src: "/api/images/zTsAAAAAAAA/"}
		img6 := model.Image{Id: 0x3BC8, Token: 0x3BCE, Src: "/api/images/zjsAAAAAAAA/"}
		imgs3 := []model.Image{img5, img6}
		assert.NotEqual(t, img11, img12)
		AssertContains(t, imgs3, img11)
		AssertContains(t, imgs3, img12)
	})
	t.Run("Positive2", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0xFAFD + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		conf := DefaultServiceConfig
		conf.TempLinks = false
		serv := New(conf, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img7, img8, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img1 := model.Image{Id: 0xFAFF, Token: 0xFAFF, Src: "/aye-and-nay/albums/_voAAAAAAAA/images/__oAAAAAAAA"}
		img2 := model.Image{Id: 0xFB00, Token: 0xFB00, Src: "/aye-and-nay/albums/_voAAAAAAAA/images/APsAAAAAAAA"}
		imgs1 := []model.Image{img1, img2}
		assert.NotEqual(t, img7, img8)
		AssertContains(t, imgs1, img7)
		AssertContains(t, imgs1, img8)
		img9, img10, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img3 := model.Image{Id: 0xFB00, Token: 0xFB00, Src: "/aye-and-nay/albums/_voAAAAAAAA/images/APsAAAAAAAA"}
		img4 := model.Image{Id: 0xFAFF, Token: 0xFAFF, Src: "/aye-and-nay/albums/_voAAAAAAAA/images/__oAAAAAAAA"}
		imgs2 := []model.Image{img3, img4}
		assert.NotEqual(t, img9, img10)
		AssertContains(t, imgs2, img9)
		AssertContains(t, imgs2, img10)
		img11, img12, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img5 := model.Image{Id: 0xFAFF, Token: 0xFAFF, Src: "/aye-and-nay/albums/_voAAAAAAAA/images/__oAAAAAAAA"}
		img6 := model.Image{Id: 0xFB00, Token: 0xFB00, Src: "/aye-and-nay/albums/_voAAAAAAAA/images/APsAAAAAAAA"}
		imgs3 := []model.Image{img5, img6}
		assert.NotEqual(t, img11, img12)
		AssertContains(t, imgs3, img11)
		AssertContains(t, imgs3, img12)
	})
	t.Run("Negative", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel)
		_, _, err := serv.Pair(ctx, 0xEB46)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestServiceImage(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0xA83F + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		f, err := serv.Image(ctx, img1.Token)
		assert.NoError(t, err)
		assert.NotNil(t, f.Reader)
		f, err = serv.Image(ctx, img2.Token)
		assert.NoError(t, err)
		assert.NotNil(t, f.Reader)
	})
	t.Run("Negative", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel)
		_, err := serv.Image(ctx, 0xE283)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}

func TestServiceVote(t *testing.T) {
	t.Run("Positive1", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0xC389 + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		err = serv.Vote(ctx, album, img1.Token, img2.Token)
		assert.NoError(t, err)
	})
	t.Run("Positive2", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0x1E58 + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		conf := DefaultServiceConfig
		conf.TempLinks = false
		serv := New(conf, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		err = serv.Vote(ctx, album, img1.Token, img2.Token)
		assert.NoError(t, err)
	})
	t.Run("Negative1", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0xE24F + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		err = serv.Vote(ctx, 0x12E6, img1.Token, img2.Token)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative2", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0xBC43 + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		_, _, err = serv.Pair(ctx, album)
		assert.NoError(t, err)
		err = serv.Vote(ctx, album, 0x1CC1, 0xF83C)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}

func TestServiceTop(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0x4DB8 + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{newQueue(0x1A01, mCache)}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		heartbeatCalc := make(chan interface{})
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2), WithHeartbeatCalc(heartbeatCalc))
		gCalc, ctxCalc := errgroup.WithContext(ctx)
		serv.StartWorkingPoolCalc(ctxCalc, gCalc)
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		err = serv.Vote(ctx, album, img1.Token, img2.Token)
		assert.NoError(t, err)
		AssertChannel(t, heartbeatCalc)
		img3, img4, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		err = serv.Vote(ctx, album, img3.Token, img4.Token)
		assert.NoError(t, err)
		AssertChannel(t, heartbeatCalc)
		imgs1, err := serv.Top(ctx, album)
		assert.NoError(t, err)
		img5 := model.Image{Id: 0x4DBA, Src: "/aye-and-nay/albums/uU0AAAAAAAA/images/uk0AAAAAAAA", Rating: 0.5, Compressed: false}
		img6 := model.Image{Id: 0x4DBB, Src: "/aye-and-nay/albums/uU0AAAAAAAA/images/u00AAAAAAAA", Rating: 0.5, Compressed: false}
		imgs2 := []model.Image{img5, img6}
		assert.Equal(t, imgs2, imgs1)
	})
	t.Run("Negative", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel)
		_, err := serv.Top(ctx, 0x83CD)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestServiceDelete(t *testing.T) {
	t.Run("Positive1", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{newPQueue(0xE3FF, mCache)}
		qDel.Monitor(ctx)
		alb1 := AlbumEmptyFactory(0x101F)
		alb1.Expires = time.Now().Add(-1 * time.Hour)
		err := mDb.SaveAlbum(ctx, alb1)
		assert.NoError(t, err)
		alb2 := AlbumEmptyFactory(0xFFBB)
		alb2.Expires = time.Now().Add(1 * time.Hour)
		err = mDb.SaveAlbum(ctx, alb2)
		assert.NoError(t, err)
		heartbeatDel := make(chan interface{})
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithHeartbeatDel(heartbeatDel))
		err = serv.CleanUp(ctx)
		assert.NoError(t, err)
		gDel, ctxDel := errgroup.WithContext(ctx)
		serv.StartWorkingPoolDel(ctxDel, gDel)
		v := AssertChannel(t, heartbeatDel)
		album, ok := v.(uint64)
		assert.True(t, ok)
		assert.Equal(t, uint64(0x101F), album)
	})
	t.Run("Positive2", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{newPQueue(0xEF3F, mCache)}
		qDel.Monitor(ctx)
		heartbeatDel := make(chan interface{})
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithHeartbeatDel(heartbeatDel))
		gDel, ctxDel := errgroup.WithContext(ctx)
		serv.StartWorkingPoolDel(ctxDel, gDel)
		files := []model.File{Png(), Png()}
		dur := 100 * time.Millisecond
		album, err := serv.Album(ctx, files, dur)
		assert.NoError(t, err)
		AssertChannel(t, heartbeatDel)
		_, err = serv.Top(ctx, album)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem(database.DefaultMemConfig)
		mCache := cache.NewMem(cache.DefaultMemConfig)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{newPQueue(0xEF3F, mCache)}
		qDel.Monitor(ctx)
		heartbeatDel := make(chan interface{})
		serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel, WithHeartbeatDel(heartbeatDel))
		gDel, ctxDel := errgroup.WithContext(ctx)
		serv.StartWorkingPoolDel(ctxDel, gDel)
		files := []model.File{Png(), Png()}
		dur := 0 * time.Second
		album, err := serv.Album(ctx, files, dur)
		assert.NoError(t, err)
		AssertNotChannel(t, heartbeatDel)
		_, err = serv.Top(ctx, album)
		assert.NoError(t, err)
	})
}

func TestServiceHealth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	comp := compressor.NewMock()
	stor := storage.NewMock()
	mDb := database.NewMem(database.DefaultMemConfig)
	mCache := cache.NewMem(cache.DefaultMemConfig)
	qCalc := &QueueCalc{}
	qCalc.Monitor(ctx)
	qComp := &QueueComp{}
	qComp.Monitor(ctx)
	qDel := &QueueDel{}
	qDel.Monitor(ctx)
	serv := New(DefaultServiceConfig, comp, stor, mDb, mCache, qCalc, qComp, qDel)
	_, err := serv.Health(ctx)
	assert.NoError(t, err)
}
