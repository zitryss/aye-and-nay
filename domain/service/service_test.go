package service

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func TestServiceAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "zcU244KtR3jJrnt9"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		ctx := context.Background()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*Queue)(nil)
		queue2 := NewQueue("TV7ZuMmhz3CDfa7n", &mem)
		pqueue := (*PQueue)(nil)
		heartbeatComp := make(chan interface{})
		serv := NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, WithRandId(fn1), WithHeartbeatComp(heartbeatComp))
		g, ctx2 := errgroup.WithContext(ctx)
		serv.StartWorkingPoolComp(ctx2, g)
		files := []model.File{Png(), Png()}
		_, err := serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		v := CheckChannel(t, heartbeatComp)
		p, ok := v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 0.5) {
			t.Error("p != 0.5")
		}
		v = CheckChannel(t, heartbeatComp)
		p, ok = v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 1) {
			t.Error("p != 1")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "wZE65QekXNTP9vpK"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		ctx := context.Background()
		heartbeatRestart := make(chan interface{})
		comp := compressor.NewFail(compressor.WithHeartbeatRestart(heartbeatRestart))
		comp.Monitor()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*Queue)(nil)
		queue2 := NewQueue("mhynV9uhnGFEV4uf", &mem)
		pqueue := (*PQueue)(nil)
		heartbeatComp := make(chan interface{})
		serv := NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, WithRandId(fn1), WithHeartbeatComp(heartbeatComp))
		g, ctx2 := errgroup.WithContext(ctx)
		serv.StartWorkingPoolComp(ctx2, g)
		files := []model.File{Png(), Png()}
		_, err := serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		v := CheckChannel(t, heartbeatComp)
		err, ok := v.(error)
		if !ok {
			t.Error("v.(type) != error")
		}
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
		files = []model.File{Png(), Png()}
		_, err = serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		v = CheckChannel(t, heartbeatComp)
		p, ok := v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 0.5) {
			t.Error("p != 0.5")
		}
		v = CheckChannel(t, heartbeatComp)
		p, ok = v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 1) {
			t.Error("p != 1")
		}
		CheckChannel(t, heartbeatRestart)
		CheckChannel(t, heartbeatRestart)
		files = []model.File{Png(), Png()}
		_, err = serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		v = CheckChannel(t, heartbeatComp)
		err, ok = v.(error)
		if !ok {
			t.Error("v.(type) != error")
		}
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
		files = []model.File{Png(), Png()}
		_, err = serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		v = CheckChannel(t, heartbeatComp)
		p, ok = v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 0.5) {
			t.Error("p != 0.5")
		}
		v = CheckChannel(t, heartbeatComp)
		p, ok = v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 1) {
			t.Error("p != 1")
		}
	})
}

func TestServicePair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "Rkur9G4z9PKtURHe"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*Queue)(nil)
		queue2 := (*Queue)(nil)
		pqueue := (*PQueue)(nil)
		serv := NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
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
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*Queue)(nil)
		queue2 := (*Queue)(nil)
		pqueue := (*PQueue)(nil)
		serv := NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue)
		_, _, err := serv.Pair(ctx, "A755jF7tvnTJrPCD")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestServiceVote(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "kh6yGRSrzXXqW9Ap"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*Queue)(nil)
		queue2 := (*Queue)(nil)
		pqueue := (*PQueue)(nil)
		serv := NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
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
		fn1 := func() func(int) (string, error) {
			id := "4UF24e4Ka9UWtEdg"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*Queue)(nil)
		queue2 := (*Queue)(nil)
		pqueue := (*PQueue)(nil)
		serv := NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
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
		fn1 := func() func(int) (string, error) {
			id := "hw9mwZyRgxBC9Xbt"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*Queue)(nil)
		queue2 := (*Queue)(nil)
		pqueue := (*PQueue)(nil)
		serv := NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
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

func TestServiceTop(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func(int) (string, error) {
			id := "L2j8Uc3z2HNLZHvJ"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		ctx := context.Background()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := NewQueue("RKvUKsDj7whcrpzA", &mem)
		queue2 := (*Queue)(nil)
		pqueue := (*PQueue)(nil)
		heartbeatCalc := make(chan interface{})
		serv := NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, WithRandId(fn1), WithRandShuffle(fn2), WithHeartbeatCalc(heartbeatCalc))
		g1, ctx1 := errgroup.WithContext(ctx)
		serv.StartWorkingPoolCalc(ctx1, g1)
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
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
		CheckChannel(t, heartbeatCalc)
		img3, img4, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		err = serv.Vote(ctx, album, img3.Token, img4.Token)
		if err != nil {
			t.Error(err)
		}
		CheckChannel(t, heartbeatCalc)
		imgs1, err := serv.Top(ctx, album)
		if err != nil {
			t.Error(err)
		}
		img5 := model.Image{Id: "L2j8Uc3z2HNLZHvJ2", Src: "/aye-and-nay/albums/L2j8Uc3z2HNLZHvJ1/images/L2j8Uc3z2HNLZHvJ2", Rating: 0.5, Compressed: false}
		img6 := model.Image{Id: "L2j8Uc3z2HNLZHvJ3", Src: "/aye-and-nay/albums/L2j8Uc3z2HNLZHvJ1/images/L2j8Uc3z2HNLZHvJ3", Rating: 0.5, Compressed: false}
		imgs2 := []model.Image{img5, img6}
		if !reflect.DeepEqual(imgs1, imgs2) {
			t.Error("imgs1 != imgs2")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		ctx := context.Background()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mem := database.NewMem()
		queue1 := (*Queue)(nil)
		queue2 := (*Queue)(nil)
		pqueue := (*PQueue)(nil)
		serv := NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue)
		_, err := serv.Top(ctx, "XXAzCcc6EHr6mpcH")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestServiceDelete(t *testing.T) {
	fn := func() time.Time {
		return time.Now()
	}
	ctx := context.Background()
	comp := compressor.NewMock()
	stor := storage.NewMock()
	mem := database.NewMem()
	queue1 := (*Queue)(nil)
	queue2 := (*Queue)(nil)
	pqueue := NewPQueue("en8wWYq2ms5Zgnw7", &mem)
	pqueue.Monitor(ctx)
	heartbeatDel := make(chan interface{})
	serv := NewService(&comp, &stor, &mem, &mem, queue1, queue2, pqueue, WithRandNow(fn), WithHeartbeatDel(heartbeatDel))
	g1, ctx1 := errgroup.WithContext(ctx)
	serv.StartWorkingPoolDel(ctx1, g1)
	files := []model.File{Png(), Png()}
	dur := 100 * time.Millisecond
	album, err := serv.Album(ctx, files, dur)
	if err != nil {
		t.Error(err)
	}
	select {
	case <-heartbeatDel:
	case <-time.After(120 * time.Millisecond):
		t.Error("<-time.After(120 * time.Millisecond)")
	}
	_, err = serv.Top(ctx, album)
	if !errors.Is(err, model.ErrAlbumNotFound) {
		t.Error(err)
	}
}
