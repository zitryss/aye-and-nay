// +build integration

package service

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	"github.com/zitryss/aye-and-nay/internal/dockertest"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/env"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
	"github.com/zitryss/aye-and-nay/pkg/rand"
)

func TestMain(m *testing.M) {
	_, err := env.Lookup("CONTINUOUS_INTEGRATION")
	if err != nil {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.Lcritical)
		docker := dockertest.New()
		docker.RunMongo()
		docker.RunRedis()
		docker.RunMinio()
		code := m.Run()
		docker.Purge()
		os.Exit(code)
	}
	code := m.Run()
	os.Exit(code)
}

func TestServiceIntegrationAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "zcU244KtR3jJrnt9"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		ctx := context.Background()
		comp := compressor.NewMock()
		stor, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		sched := NewScheduler("VK4dE8CgS82B8yC7", &redis)
		serv := NewService(&comp, &stor, &mongo, &redis, &sched)
		files := [][]byte{nil, nil}
		_, err = serv.Album(ctx, files)
		if err != nil {
			t.Error(err)
		}
		found, err := serv.Exists(ctx, "zcU244KtR3jJrnt91")
		if err != nil {
			t.Error(err)
		}
		if !found {
			t.Error("album not found")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "usG3VzbLvRAm2k2y"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		ctx := context.Background()
		comp := compressor.NewFail()
		stor, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		sched := NewScheduler("DgWwCAxe2JUpJbHt", &redis)
		serv := NewService(&comp, &stor, &mongo, &redis, &sched)
		files := [][]byte{nil, nil}
		_, err = serv.Album(ctx, files)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
		_, err = serv.Album(ctx, files)
		if err != nil {
			t.Error(err)
		}
		found, err := serv.Exists(ctx, "usG3VzbLvRAm2k2y4")
		if err != nil {
			t.Error(err)
		}
		if !found {
			t.Error("album not found")
		}
	})
}

func TestServiceIntegrationPair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "Rkur9G4z9PKtURHe"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		rand.Shuffle = func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		sched := NewScheduler("766fFt8nuJ5qRek2", &redis)
		serv := NewService(&comp, &stor, &mongo, &redis, &sched)
		files := [][]byte{nil, nil}
		album, err := serv.Album(ctx, files)
		if err != nil {
			t.Error(err)
		}
		img7, img8, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		img1 := model.Image{Id: "Rkur9G4z9PKtURHe2", Token: "Rkur9G4z9PKtURHe4", Src: "/aye-and-nay/albums/Rkur9G4z9PKtURHe1/images/Rkur9G4z9PKtURHe2"}
		img2 := model.Image{Id: "Rkur9G4z9PKtURHe3", Token: "Rkur9G4z9PKtURHe5", Src: "/aye-and-nay/albums/Rkur9G4z9PKtURHe1/images/Rkur9G4z9PKtURHe3"}
		imgs1 := []model.Image{img1, img2}
		if reflect.DeepEqual(img7, img8) {
			t.Error("img7 == img8")
		}
		if !IsIn(img7, imgs1) {
			t.Error("img7 is not in imgs")
		}
		if !IsIn(img8, imgs1) {
			t.Error("img8 is not in imgs")
		}
		img9, img10, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		img3 := model.Image{Id: "Rkur9G4z9PKtURHe3", Token: "Rkur9G4z9PKtURHe6", Src: "/aye-and-nay/albums/Rkur9G4z9PKtURHe1/images/Rkur9G4z9PKtURHe3"}
		img4 := model.Image{Id: "Rkur9G4z9PKtURHe2", Token: "Rkur9G4z9PKtURHe7", Src: "/aye-and-nay/albums/Rkur9G4z9PKtURHe1/images/Rkur9G4z9PKtURHe2"}
		imgs2 := []model.Image{img3, img4}
		if reflect.DeepEqual(img9, img10) {
			t.Error("img9 == img10")
		}
		if !IsIn(img9, imgs2) {
			t.Error("img9 is not in imgs")
		}
		if !IsIn(img10, imgs2) {
			t.Error("img10 is not in imgs")
		}
		img11, img12, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		img5 := model.Image{Id: "Rkur9G4z9PKtURHe2", Token: "Rkur9G4z9PKtURHe8", Src: "/aye-and-nay/albums/Rkur9G4z9PKtURHe1/images/Rkur9G4z9PKtURHe2"}
		img6 := model.Image{Id: "Rkur9G4z9PKtURHe3", Token: "Rkur9G4z9PKtURHe9", Src: "/aye-and-nay/albums/Rkur9G4z9PKtURHe1/images/Rkur9G4z9PKtURHe3"}
		imgs3 := []model.Image{img5, img6}
		if reflect.DeepEqual(img11, img12) {
			t.Error("img11 == img12")
		}
		if !IsIn(img11, imgs3) {
			t.Error("img11 is not in imgs")
		}
		if !IsIn(img12, imgs3) {
			t.Error("img12 is not in imgs")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		ctx := context.Background()
		comp := compressor.NewMock()
		stor, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		sched := NewScheduler("kRneghVzdmtScFYG", &redis)
		serv := NewService(&comp, &stor, &mongo, &redis, &sched)
		_, _, err = serv.Pair(ctx, "A755jF7tvnTJrPCD")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestServiceIntegrationVote(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "kh6yGRSrzXXqW9Ap"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		rand.Shuffle = func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		sched := NewScheduler("8eDkyz293xggaUpr", &redis)
		serv := NewService(&comp, &stor, &mongo, &redis, &sched)
		files := [][]byte{nil, nil}
		album, err := serv.Album(ctx, files)
		if err != nil {
			t.Error(err)
		}
		img1, img2, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		err = serv.Vote(ctx, album, img1.Token, img2.Token)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "4UF24e4Ka9UWtEdg"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		rand.Shuffle = func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		sched := NewScheduler("b8mKspbYz5FjQ7Mf", &redis)
		serv := NewService(&comp, &stor, &mongo, &redis, &sched)
		files := [][]byte{nil, nil}
		album, err := serv.Album(ctx, files)
		if err != nil {
			t.Error(err)
		}
		img1, img2, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		err = serv.Vote(ctx, "tHwPdF76b3DahJrP", img1.Token, img2.Token)
		if !errors.Is(err, model.ErrTokenNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "hw9mwZyRgxBC9Xbt"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		rand.Shuffle = func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		sched := NewScheduler("nRQynzFJvPvcRZUt", &redis)
		serv := NewService(&comp, &stor, &mongo, &redis, &sched)
		files := [][]byte{nil, nil}
		album, err := serv.Album(ctx, files)
		if err != nil {
			t.Error(err)
		}
		_, _, err = serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		err = serv.Vote(ctx, album, "h9zY3PqD3ng7MJxk", "mhVPPxW2GmqLBZwL")
		if !errors.Is(err, model.ErrTokenNotFound) {
			t.Error(err)
		}
	})
}

func TestServiceIntegrationTop(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "L2j8Uc3z2HNLZHvJ"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		rand.Shuffle = func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		sched := NewScheduler("RKvUKsDj7whcrpzA", &redis)
		serv := NewService(&comp, &stor, &mongo, &redis, &sched)
		g, ctx := errgroup.WithContext(ctx)
		heartbeat := make(chan struct{})
		serv.StartWorkingPool(ctx, g, heartbeat)
		files := [][]byte{nil, nil}
		album, err := serv.Album(ctx, files)
		if err != nil {
			t.Error(err)
		}
		img1, img2, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		err = serv.Vote(ctx, album, img1.Token, img2.Token)
		if err != nil {
			t.Error(err)
		}
		<-heartbeat
		img3, img4, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		err = serv.Vote(ctx, album, img3.Token, img4.Token)
		if err != nil {
			t.Error(err)
		}
		<-heartbeat
		imgs1, err := serv.Top(ctx, album)
		if err != nil {
			t.Error(err)
		}
		img5 := model.Image{Id: "L2j8Uc3z2HNLZHvJ2", Src: "/aye-and-nay/albums/L2j8Uc3z2HNLZHvJ1/images/L2j8Uc3z2HNLZHvJ2", Rating: 0.5}
		img6 := model.Image{Id: "L2j8Uc3z2HNLZHvJ3", Src: "/aye-and-nay/albums/L2j8Uc3z2HNLZHvJ1/images/L2j8Uc3z2HNLZHvJ3", Rating: 0.5}
		imgs2 := []model.Image{img5, img6}
		if !reflect.DeepEqual(imgs1, imgs2) {
			t.Error("imgs1 != imgs2")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		ctx := context.Background()
		comp := compressor.NewMock()
		stor, err := storage.NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		mongo, err := database.NewMongo()
		if err != nil {
			t.Fatal(err)
		}
		redis, err := database.NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		sched := NewScheduler("YNhuDMs3jKpVBM7E", &redis)
		serv := NewService(&comp, &stor, &mongo, &redis, &sched)
		_, err = serv.Top(ctx, "XXAzCcc6EHr6mpcH")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}
