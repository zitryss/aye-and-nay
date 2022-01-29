//go:build integration

package database

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestMongoAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(0x6CC4)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		edgs, err := mongo.GetEdges(context.Background(), 0x6CC4)
		assert.NoError(t, err)
		assert.Equal(t, alb.Edges, edgs)
	})
	t.Run("Negative1", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumFullFactory(0xA566)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		alb = AlbumFullFactory(0xA566)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.ErrorIs(t, err, domain.ErrAlbumAlreadyExists)
	})
	t.Run("Negative2", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		_, err = mongo.GetImagesIds(context.Background(), 0xA9B4)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative3", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		_, err = mongo.GetEdges(context.Background(), 0x3F1E)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestMongoCount(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(0x746C)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		n, err := mongo.CountImages(context.Background(), 0x746C)
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		n, err = mongo.CountImagesCompressed(context.Background(), 0x746C)
		assert.NoError(t, err)
		assert.Equal(t, 0, n)
		err = mongo.UpdateCompressionStatus(context.Background(), 0x746C, 0x3E3D)
		assert.NoError(t, err)
		n, err = mongo.CountImages(context.Background(), 0x746C)
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		n, err = mongo.CountImagesCompressed(context.Background(), 0x746C)
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
		err = mongo.UpdateCompressionStatus(context.Background(), 0x746C, 0xB399)
		assert.NoError(t, err)
		n, err = mongo.CountImagesCompressed(context.Background(), 0x746C)
		assert.NoError(t, err)
		assert.Equal(t, 2, n)
	})
	t.Run("Negative1", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(0x99DF)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		err = mongo.UpdateCompressionStatus(context.Background(), 0x99DF, 0x3E3D)
		assert.NoError(t, err)
		err = mongo.UpdateCompressionStatus(context.Background(), 0x99DF, 0x3E3D)
		assert.NoError(t, err)
		n, err := mongo.CountImagesCompressed(context.Background(), 0x99DF)
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
	})
	t.Run("Negative2", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		_, err = mongo.CountImages(context.Background(), 0xF256)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative3", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		_, err = mongo.CountImagesCompressed(context.Background(), 0xC52A)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative4", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		err = mongo.UpdateCompressionStatus(context.Background(), 0xF73E, 0x3E3D)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative5", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(0xDF75)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		err = mongo.UpdateCompressionStatus(context.Background(), 0xDF75, 0xE7A4)
		assert.ErrorIs(t, err, domain.ErrImageNotFound)
	})
}

func TestMongoImage(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(0xB0C4)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		src, err := mongo.GetImageSrc(context.Background(), 0xB0C4, 0x51DE)
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/xLAAAAAAAAA/images/3lEAAAAAAAA", src)
	})
	t.Run("Negative1", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		_, err = mongo.GetImageSrc(context.Background(), 0x12EE, 0x51DE)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative2", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(0xD585)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		_, err = mongo.GetImageSrc(context.Background(), 0xD585, 0xDA30)
		assert.ErrorIs(t, err, domain.ErrImageNotFound)
	})
}

func TestMongoVote(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumFullFactory(0x4D76)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		err = mongo.SaveVote(context.Background(), 0x4D76, 0xDA2A, 0xDA52)
		assert.NoError(t, err)
		err = mongo.SaveVote(context.Background(), 0x4D76, 0xDA2A, 0xDA52)
		assert.NoError(t, err)
		edgs, err := mongo.GetEdges(context.Background(), 0x4D76)
		assert.NoError(t, err)
		assert.Equal(t, 2, edgs[0xDA2A][0xDA52])
	})
	t.Run("Negative", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		err = mongo.SaveVote(context.Background(), 0x1FAD, 0x84E6, 0x308E)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestMongoSort(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumFullFactory(0x5A96)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		imgs1, err := mongo.GetImagesOrdered(context.Background(), 0x5A96)
		assert.NoError(t, err)
		img1 := model.Image{Id: 0x51DE, Src: "/aye-and-nay/albums/lloAAAAAAAA/images/3lEAAAAAAAA", Rating: 0.77920413}
		img2 := model.Image{Id: 0x3E3D, Src: "/aye-and-nay/albums/lloAAAAAAAA/images/PT4AAAAAAAA", Rating: 0.48954984}
		img3 := model.Image{Id: 0xDA2A, Src: "/aye-and-nay/albums/lloAAAAAAAA/images/KtoAAAAAAAA", Rating: 0.41218211}
		img4 := model.Image{Id: 0xB399, Src: "/aye-and-nay/albums/lloAAAAAAAA/images/mbMAAAAAAAA", Rating: 0.19186324}
		img5 := model.Image{Id: 0xDA52, Src: "/aye-and-nay/albums/lloAAAAAAAA/images/UtoAAAAAAAA", Rating: 0.13278389}
		imgs2 := []model.Image{img1, img2, img3, img4, img5}
		assert.Equal(t, imgs2, imgs1)
	})
	t.Run("Negative", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		_, err = mongo.GetImagesOrdered(context.Background(), 0x66BE)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestMongoRatings(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumFullFactory(0x4E54)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
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
		err = mongo.UpdateRatings(context.Background(), 0x4E54, vector)
		assert.NoError(t, err)
		imgs2, err := mongo.GetImagesOrdered(context.Background(), 0x4E54)
		assert.NoError(t, err)
		sort.Slice(imgs1, func(i, j int) bool { return imgs1[i].Rating > imgs1[j].Rating })
		assert.Equal(t, imgs1, imgs2)
	})
	t.Run("Negative", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
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
		err = mongo.UpdateRatings(context.Background(), 0xA293, vector)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestMongoDelete(t *testing.T) {
	t.Run("Positive1", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(0x748C)
		_, err = mongo.CountImages(context.Background(), 0x748C)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		n, err := mongo.CountImages(context.Background(), 0x748C)
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		albums, err := mongo.AlbumsToBeDeleted(context.Background())
		assert.NoError(t, err)
		assert.Len(t, albums, 0)
		err = mongo.DeleteAlbum(context.Background(), 0x748C)
		assert.NoError(t, err)
		_, err = mongo.CountImages(context.Background(), 0x748C)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Positive2", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(0x7B43)
		alb.Expires = time.Now().Add(-1 * time.Hour)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		albums, err := mongo.AlbumsToBeDeleted(context.Background())
		assert.NoError(t, err)
		assert.True(t, len(albums) == 1 && albums[0].Id == alb.Id && !albums[0].Expires.IsZero())
		err = mongo.DeleteAlbum(context.Background(), 0x7B43)
		assert.NoError(t, err)
		_, err = mongo.CountImages(context.Background(), 0x7B43)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
		t.Cleanup(func() { _ = mongo.DeleteAlbum(context.Background(), 0x7B43) })
	})
	t.Run("Negative", func(t *testing.T) {
		mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(0x608C)
		err = mongo.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		err = mongo.DeleteAlbum(context.Background(), 0xB7FF)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestMongoLru(t *testing.T) {
	mongo, err := NewMongo(context.Background(), DefaultMongoConfig)
	require.NoError(t, err)
	alb1 := AlbumEmptyFactory(0x36FC)
	err = mongo.SaveAlbum(context.Background(), alb1)
	assert.NoError(t, err)
	alb2 := AlbumEmptyFactory(0xB020)
	err = mongo.SaveAlbum(context.Background(), alb2)
	assert.NoError(t, err)
	edgs, err := mongo.GetEdges(context.Background(), 0x36FC)
	assert.NoError(t, err)
	assert.Equal(t, alb1.Edges, edgs)
}
