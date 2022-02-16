package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	. "github.com/zitryss/aye-and-nay/internal/generator"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestServiceIntegrationAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, _ := GenId()
		idQ, _ := GenId()
		fnId := func() func() (uint64, error) {
			return func() (uint64, error) {
				return id(), nil
			}
		}()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{newQueue(idQ(), redis)}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		heartbeatComp := make(chan interface{})
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithRandId(fnId), WithHeartbeatComp(heartbeatComp))
		gComp, ctxComp := errgroup.WithContext(ctx)
		serv.StartWorkingPoolComp(ctxComp, gComp)
		files := []model.File{Png(), Png()}
		_, err = serv.Album(ctx, files, 0*time.Millisecond)
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
		id, _ := GenId()
		idQ, _ := GenId()
		fnId := func() func() (uint64, error) {
			return func() (uint64, error) {
				return id(), nil
			}
		}()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		heartbeatRestart := make(chan interface{})
		comp := compressor.NewShortpixel(compressor.DefaultShortpixelConfig, compressor.WithHeartbeatRestart(heartbeatRestart))
		comp.Monitor(ctx)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{newQueue(idQ(), redis)}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		heartbeatComp := make(chan interface{})
		serv := New(DefaultServiceConfig, comp, minio, mongo, redis, qCalc, qComp, qDel, WithRandId(fnId), WithHeartbeatComp(heartbeatComp))
		gComp, ctxComp := errgroup.WithContext(ctx)
		serv.StartWorkingPoolComp(ctxComp, gComp)
		files := []model.File{Png(), Png()}
		_, err = serv.Album(ctx, files, 0*time.Millisecond)
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

func TestServiceIntegrationPair(t *testing.T) {
	t.Run("Positive1", func(t *testing.T) {
		id, ids := GenId()
		fnId := func() func() (uint64, error) {
			return func() (uint64, error) {
				return id(), nil
			}
		}()
		fnS := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithRandId(fnId), WithRandShuffle(fnS))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img7, img8, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img1 := model.Image{Id: ids.Uint64(1), Token: ids.Uint64(3), Src: "/api/images/" + ids.Base64(3) + "/"}
		img2 := model.Image{Id: ids.Uint64(2), Token: ids.Uint64(4), Src: "/api/images/" + ids.Base64(4) + "/"}
		imgs1 := []model.Image{img1, img2}
		assert.NotEqual(t, img7, img8)
		assert.Contains(t, imgs1, img7)
		assert.Contains(t, imgs1, img8)
		img9, img10, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img3 := model.Image{Id: ids.Uint64(2), Token: ids.Uint64(5), Src: "/api/images/" + ids.Base64(5) + "/"}
		img4 := model.Image{Id: ids.Uint64(1), Token: ids.Uint64(6), Src: "/api/images/" + ids.Base64(6) + "/"}
		imgs2 := []model.Image{img3, img4}
		assert.NotEqual(t, img9, img10)
		assert.Contains(t, imgs2, img9)
		assert.Contains(t, imgs2, img10)
		img11, img12, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img5 := model.Image{Id: ids.Uint64(1), Token: ids.Uint64(7), Src: "/api/images/" + ids.Base64(7) + "/"}
		img6 := model.Image{Id: ids.Uint64(2), Token: ids.Uint64(8), Src: "/api/images/" + ids.Base64(8) + "/"}
		imgs3 := []model.Image{img5, img6}
		assert.NotEqual(t, img11, img12)
		assert.Contains(t, imgs3, img11)
		assert.Contains(t, imgs3, img12)
	})
	t.Run("Positive2", func(t *testing.T) {
		id, ids := GenId()
		fnId := func() func() (uint64, error) {
			return func() (uint64, error) {
				return id(), nil
			}
		}()
		fnS := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		conf := DefaultServiceConfig
		conf.TempLinks = false
		serv := New(conf, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithRandId(fnId), WithRandShuffle(fnS))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img7, img8, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img1 := model.Image{Id: ids.Uint64(1), Token: ids.Uint64(1), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(1)}
		img2 := model.Image{Id: ids.Uint64(2), Token: ids.Uint64(2), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(2)}
		imgs1 := []model.Image{img1, img2}
		assert.NotEqual(t, img7, img8)
		assert.Contains(t, imgs1, img7)
		assert.Contains(t, imgs1, img8)
		img9, img10, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img3 := model.Image{Id: ids.Uint64(2), Token: ids.Uint64(2), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(2)}
		img4 := model.Image{Id: ids.Uint64(1), Token: ids.Uint64(1), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(1)}
		imgs2 := []model.Image{img3, img4}
		assert.NotEqual(t, img9, img10)
		assert.Contains(t, imgs2, img9)
		assert.Contains(t, imgs2, img10)
		img11, img12, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		img5 := model.Image{Id: ids.Uint64(1), Token: ids.Uint64(1), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(1)}
		img6 := model.Image{Id: ids.Uint64(2), Token: ids.Uint64(2), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(2)}
		imgs3 := []model.Image{img5, img6}
		assert.NotEqual(t, img11, img12)
		assert.Contains(t, imgs3, img11)
		assert.Contains(t, imgs3, img12)
	})
	t.Run("Negative", func(t *testing.T) {
		id, _ := GenId()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel)
		_, _, err = serv.Pair(ctx, id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestServiceIntegrationImage(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, _ := GenId()
		fnId := func() func() (uint64, error) {
			return func() (uint64, error) {
				return id(), nil
			}
		}()
		fnS := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithRandId(fnId), WithRandShuffle(fnS))
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
		id, _ := GenId()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel)
		_, err = serv.Image(ctx, id())
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}

func TestServiceIntegrationVote(t *testing.T) {
	t.Run("Positive1", func(t *testing.T) {
		id, _ := GenId()
		fnId := func() func() (uint64, error) {
			return func() (uint64, error) {
				return id(), nil
			}
		}()
		fnS := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithRandId(fnId), WithRandShuffle(fnS))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		err = serv.Vote(ctx, album, img1.Token, img2.Token)
		assert.NoError(t, err)
	})
	t.Run("Positive2", func(t *testing.T) {
		id, _ := GenId()
		fnId := func() func() (uint64, error) {
			return func() (uint64, error) {
				return id(), nil
			}
		}()
		fnS := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		conf := DefaultServiceConfig
		conf.TempLinks = false
		serv := New(conf, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithRandId(fnId), WithRandShuffle(fnS))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		err = serv.Vote(ctx, album, img1.Token, img2.Token)
		assert.NoError(t, err)
	})
	t.Run("Negative1", func(t *testing.T) {
		id, _ := GenId()
		fnId := func() func() (uint64, error) {
			return func() (uint64, error) {
				return id(), nil
			}
		}()
		fnS := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithRandId(fnId), WithRandShuffle(fnS))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := serv.Pair(ctx, album)
		assert.NoError(t, err)
		err = serv.Vote(ctx, id(), img1.Token, img2.Token)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative2", func(t *testing.T) {
		id, _ := GenId()
		fnId := func() func() (uint64, error) {
			return func() (uint64, error) {
				return id(), nil
			}
		}()
		fnS := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithRandId(fnId), WithRandShuffle(fnS))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		_, _, err = serv.Pair(ctx, album)
		assert.NoError(t, err)
		err = serv.Vote(ctx, album, id(), id())
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}

func TestServiceIntegrationTop(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, ids := GenId()
		idQ, _ := GenId()
		fnId := func() func() (uint64, error) {
			return func() (uint64, error) {
				return id(), nil
			}
		}()
		fnS := func(n int, swap func(i int, j int)) {}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{newQueue(idQ(), redis)}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		heartbeatCalc := make(chan interface{})
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithRandId(fnId), WithRandShuffle(fnS), WithHeartbeatCalc(heartbeatCalc))
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
		img5 := model.Image{Id: ids.Uint64(1), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(1), Rating: 0.5, Compressed: false}
		img6 := model.Image{Id: ids.Uint64(2), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(2), Rating: 0.5, Compressed: false}
		imgs2 := []model.Image{img5, img6}
		assert.Equal(t, imgs2, imgs1)
	})
	t.Run("Negative", func(t *testing.T) {
		id, _ := GenId()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel)
		_, err = serv.Top(ctx, id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestServiceIntegrationDelete(t *testing.T) {
	t.Run("Positive1", func(t *testing.T) {
		id1, ids1 := GenId()
		id2, ids2 := GenId()
		idPQ, _ := GenId()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{newPQueue(idPQ(), redis)}
		qDel.Monitor(ctx)
		alb1 := AlbumFactory(id1, ids1)
		alb1.Expires = time.Now().Add(-1 * time.Hour)
		err = mongo.SaveAlbum(ctx, alb1)
		assert.NoError(t, err)
		alb2 := AlbumFactory(id2, ids2)
		alb2.Expires = time.Now().Add(1 * time.Hour)
		err = mongo.SaveAlbum(ctx, alb2)
		assert.NoError(t, err)
		heartbeatDel := make(chan interface{})
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithHeartbeatDel(heartbeatDel))
		err = serv.CleanUp(ctx)
		assert.NoError(t, err)
		gDel, ctxDel := errgroup.WithContext(ctx)
		serv.StartWorkingPoolDel(ctxDel, gDel)
		v := AssertChannel(t, heartbeatDel)
		album, ok := v.(uint64)
		assert.True(t, ok)
		assert.Equal(t, ids1.Uint64(0), album)
		t.Cleanup(func() {
			_ = mongo.DeleteAlbum(context.Background(), ids1.Uint64(0))
			_ = mongo.DeleteAlbum(context.Background(), ids2.Uint64(0))
		})
	})
	t.Run("Positive2", func(t *testing.T) {
		idPQ, _ := GenId()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{newPQueue(idPQ(), redis)}
		qDel.Monitor(ctx)
		heartbeatDel := make(chan interface{})
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithHeartbeatDel(heartbeatDel))
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
		idPQ, _ := GenId()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
		require.NoError(t, err)
		minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
		require.NoError(t, err)
		mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
		require.NoError(t, err)
		redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
		require.NoError(t, err)
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{newPQueue(idPQ(), redis)}
		qDel.Monitor(ctx)
		heartbeatDel := make(chan interface{})
		serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel, WithHeartbeatDel(heartbeatDel))
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

func TestServiceIntegrationHealth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	imaginary, err := compressor.NewImaginary(ctx, compressor.DefaultImaginaryConfig)
	require.NoError(t, err)
	minio, err := storage.NewMinio(ctx, storage.DefaultMinioConfig)
	require.NoError(t, err)
	mongo, err := database.NewMongo(ctx, database.DefaultMongoConfig)
	require.NoError(t, err)
	redis, err := cache.NewRedis(ctx, cache.DefaultRedisConfig)
	require.NoError(t, err)
	qCalc := &QueueCalc{}
	qCalc.Monitor(ctx)
	qComp := &QueueComp{}
	qComp.Monitor(ctx)
	qDel := &QueueDel{}
	qDel.Monitor(ctx)
	serv := New(DefaultServiceConfig, imaginary, minio, mongo, redis, qCalc, qComp, qDel)
	_, err = serv.Health(ctx)
	assert.NoError(t, err)
}
