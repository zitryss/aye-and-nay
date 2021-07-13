package database

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/zitryss/aye-and-nay/domain/model"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func TestBadgerAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumEmptyFactory(0x6CC4)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		edgs, err := badger.GetEdges(context.Background(), 0x6CC4)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(edgs, alb.Edges) {
			t.Error("edgs != alb.GetEdges")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumFullFactory(0xA566)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		alb = AlbumFullFactory(0xA566)
		err = badger.SaveAlbum(context.Background(), alb)
		if !errors.Is(err, model.ErrAlbumAlreadyExists) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		_, err = badger.GetImagesIds(context.Background(), 0xA9B4)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		_, err = badger.GetEdges(context.Background(), 0x3F1E)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestBadgerCount(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumEmptyFactory(0x746C)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		n, err := badger.CountImages(context.Background(), 0x746C)
		if err != nil {
			t.Error(err)
		}
		if n != 5 {
			t.Error("n != 5")
		}
		n, err = badger.CountImagesCompressed(context.Background(), 0x746C)
		if err != nil {
			t.Error(err)
		}
		if n != 0 {
			t.Error("n != 0")
		}
		err = badger.UpdateCompressionStatus(context.Background(), 0x746C, 0x3E3D)
		if err != nil {
			t.Error(err)
		}
		n, err = badger.CountImages(context.Background(), 0x746C)
		if err != nil {
			t.Error(err)
		}
		if n != 5 {
			t.Error("n != 5")
		}
		n, err = badger.CountImagesCompressed(context.Background(), 0x746C)
		if err != nil {
			t.Error(err)
		}
		if n != 1 {
			t.Error("n != 1")
		}
		err = badger.UpdateCompressionStatus(context.Background(), 0x746C, 0xB399)
		if err != nil {
			t.Error(err)
		}
		n, err = badger.CountImagesCompressed(context.Background(), 0x746C)
		if err != nil {
			t.Error(err)
		}
		if n != 2 {
			t.Error("n != 2")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumEmptyFactory(0x99DF)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		err = badger.UpdateCompressionStatus(context.Background(), 0x99DF, 0x3E3D)
		if err != nil {
			t.Error(err)
		}
		err = badger.UpdateCompressionStatus(context.Background(), 0x99DF, 0x3E3D)
		if err != nil {
			t.Error(err)
		}
		n, err := badger.CountImagesCompressed(context.Background(), 0x99DF)
		if err != nil {
			t.Error(err)
		}
		if n != 1 {
			t.Error("n != 1")
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		_, err = badger.CountImages(context.Background(), 0xF256)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		_, err = badger.CountImagesCompressed(context.Background(), 0xC52A)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative4", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		err = badger.UpdateCompressionStatus(context.Background(), 0xF73E, 0x3E3D)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative5", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumEmptyFactory(0xDF75)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		err = badger.UpdateCompressionStatus(context.Background(), 0xDF75, 0xE7A4)
		if !errors.Is(err, model.ErrImageNotFound) {
			t.Error(err)
		}
	})
}

func TestBadgerImage(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumEmptyFactory(0xB0C4)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		src, err := badger.GetImageSrc(context.Background(), 0xB0C4, 0x51DE)
		if err != nil {
			t.Error(err)
		}
		if src != "/aye-and-nay/albums/xLAAAAAAAAA/images/3lEAAAAAAAA" {
			t.Error("src != \"/aye-and-nay/albums/xLAAAAAAAAA/images/3lEAAAAAAAA\"")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		_, err = badger.GetImageSrc(context.Background(), 0x12EE, 0x51DE)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumEmptyFactory(0xD585)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		_, err = badger.GetImageSrc(context.Background(), 0xD585, 0xDA30)
		if !errors.Is(err, model.ErrImageNotFound) {
			t.Error(err)
		}
	})
}

func TestBadgerVote(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumFullFactory(0x4D76)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		err = badger.SaveVote(context.Background(), 0x4D76, 0xDA2A, 0xDA52)
		if err != nil {
			t.Error(err)
		}
		err = badger.SaveVote(context.Background(), 0x4D76, 0xDA2A, 0xDA52)
		if err != nil {
			t.Error(err)
		}
		edgs, err := badger.GetEdges(context.Background(), 0x4D76)
		if err != nil {
			t.Error(err)
		}
		if edgs[0xDA2A][0xDA52] != 2 {
			t.Error("edgs[imageFrom][imageTo] != 2")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		err = badger.SaveVote(context.Background(), 0x1FAD, 0x84E6, 0x308E)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestBadgerSort(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumFullFactory(0x5A96)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		imgs1, err := badger.GetImagesOrdered(context.Background(), 0x5A96)
		if err != nil {
			t.Error(err)
		}
		img1 := model.Image{Id: 0x51DE, Src: "/aye-and-nay/albums/lloAAAAAAAA/images/3lEAAAAAAAA", Rating: 0.77920413}
		img2 := model.Image{Id: 0x3E3D, Src: "/aye-and-nay/albums/lloAAAAAAAA/images/PT4AAAAAAAA", Rating: 0.48954984}
		img3 := model.Image{Id: 0xDA2A, Src: "/aye-and-nay/albums/lloAAAAAAAA/images/KtoAAAAAAAA", Rating: 0.41218211}
		img4 := model.Image{Id: 0xB399, Src: "/aye-and-nay/albums/lloAAAAAAAA/images/mbMAAAAAAAA", Rating: 0.19186324}
		img5 := model.Image{Id: 0xDA52, Src: "/aye-and-nay/albums/lloAAAAAAAA/images/UtoAAAAAAAA", Rating: 0.13278389}
		imgs2 := []model.Image{img1, img2, img3, img4, img5}
		if !reflect.DeepEqual(imgs1, imgs2) {
			t.Error("imgs1 != imgs2")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		_, err = badger.GetImagesOrdered(context.Background(), 0x66BE)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestBadgerRatings(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumFullFactory(0x4E54)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		img1 := model.Image{Id: 0x3E3D, Src: "/aye-and-nay/albums/VE4AAAAAAAA/images/PT4AAAAAAAA", Rating: 0.54412788}
		img2 := model.Image{Id: 0xB399, Src: "/aye-and-nay/albums/VE4AAAAAAAA/images/mbMAAAAAAAA", Rating: 0.32537162}
		img3 := model.Image{Id: 0xDA2A, Src: "/aye-and-nay/albums/VE4AAAAAAAA/images/KtoAAAAAAAA", Rating: 0.43185491}
		img4 := model.Image{Id: 0x51DE, Src: "/aye-and-nay/albums/VE4AAAAAAAA/images/3lEAAAAAAAA", Rating: 0.57356209}
		img5 := model.Image{Id: 0xDA52, Src: "/aye-and-nay/albums/VE4AAAAAAAA/images/UtoAAAAAAAA", Rating: 0.61438023}
		imgs1 := []model.Image{img1, img2, img3, img4, img5}
		vector := map[uint64]float64{}
		vector[img1.Id] = img1.Rating
		vector[img2.Id] = img2.Rating
		vector[img3.Id] = img3.Rating
		vector[img4.Id] = img4.Rating
		vector[img5.Id] = img5.Rating
		err = badger.UpdateRatings(context.Background(), 0x4E54, vector)
		if err != nil {
			t.Error(err)
		}
		imgs2, err := badger.GetImagesOrdered(context.Background(), 0x4E54)
		if err != nil {
			t.Error(err)
		}
		sort.Slice(imgs1, func(i, j int) bool { return imgs1[i].Rating > imgs1[j].Rating })
		if !reflect.DeepEqual(imgs1, imgs2) {
			t.Error("imgs1 != imgs2")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		img1 := model.Image{Id: 0x3E3D, Src: "/aye-and-nay/albums/k6IAAAAAAAA/images/PT4AAAAAAAA", Rating: 0.54412788}
		img2 := model.Image{Id: 0xB399, Src: "/aye-and-nay/albums/k6IAAAAAAAA/images/mbMAAAAAAAA", Rating: 0.32537162}
		img3 := model.Image{Id: 0xDA2A, Src: "/aye-and-nay/albums/k6IAAAAAAAA/images/KtoAAAAAAAA", Rating: 0.43185491}
		img4 := model.Image{Id: 0x51DE, Src: "/aye-and-nay/albums/k6IAAAAAAAA/images/3lEAAAAAAAA", Rating: 0.57356209}
		img5 := model.Image{Id: 0xDA52, Src: "/aye-and-nay/albums/k6IAAAAAAAA/images/UtoAAAAAAAA", Rating: 0.61438023}
		vector := map[uint64]float64{}
		vector[img1.Id] = img1.Rating
		vector[img2.Id] = img2.Rating
		vector[img3.Id] = img3.Rating
		vector[img4.Id] = img4.Rating
		vector[img5.Id] = img5.Rating
		err = badger.UpdateRatings(context.Background(), 0xA293, vector)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestBadgerDelete(t *testing.T) {
	t.Run("Positive1", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumEmptyFactory(0x748C)
		_, err = badger.CountImages(context.Background(), 0x748C)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		n, err := badger.CountImages(context.Background(), 0x748C)
		if err != nil {
			t.Error(err)
		}
		if n != 5 {
			t.Error("n != 5")
		}
		albums, err := badger.AlbumsToBeDeleted(context.Background())
		if err != nil {
			t.Error(err)
		}
		if albums != nil {
			t.Error("albums != nil")
		}
		err = badger.DeleteAlbum(context.Background(), 0x748C)
		if err != nil {
			t.Error(err)
		}
		_, err = badger.CountImages(context.Background(), 0x748C)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Positive2", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumEmptyFactory(0x7B43)
		alb.Expires = time.Now().Add(-1 * time.Hour)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		albums, err := badger.AlbumsToBeDeleted(context.Background())
		if err != nil {
			t.Error(err)
		}
		if !(len(albums) == 1 && albums[0].Id == alb.Id && !albums[0].Expires.IsZero()) {
			t.Error("!(len(albums) == 1 && albums[0].Id == alb.Id && !albums[0].Expires.IsZero())")
		}
		err = badger.DeleteAlbum(context.Background(), 0x7B43)
		if err != nil {
			t.Error(err)
		}
		_, err = badger.CountImages(context.Background(), 0x7B43)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative", func(t *testing.T) {
		badger, err := NewBadger(inMemory)
		if err != nil {
			t.Fatal(err)
		}
		alb := AlbumEmptyFactory(0x608C)
		err = badger.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		err = badger.DeleteAlbum(context.Background(), 0xB7FF)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestBadgerLru(t *testing.T) {
	badger, err := NewBadger(inMemory)
	if err != nil {
		t.Fatal(err)
	}
	alb1 := AlbumEmptyFactory(0X36FC)
	err = badger.SaveAlbum(context.Background(), alb1)
	if err != nil {
		t.Error(err)
	}
	alb2 := AlbumEmptyFactory(0XB020)
	err = badger.SaveAlbum(context.Background(), alb2)
	if err != nil {
		t.Error(err)
	}
	edgs, err := badger.GetEdges(context.Background(), 0X36FC)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(edgs, alb1.Edges) {
		t.Error("edgs != alb1.GetEdges")
	}
}
