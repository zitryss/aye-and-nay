//go:build unit

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
	. "github.com/zitryss/aye-and-nay/internal/generator"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestBadgerAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(id, ids)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		edgs, err := badger.GetEdges(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, alb.Edges, edgs)
	})
	t.Run("Negative1", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumFullFactory(id, ids)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.ErrorIs(t, err, domain.ErrAlbumAlreadyExists)
	})
	t.Run("Negative2", func(t *testing.T) {
		id, _ := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		_, err = badger.GetImagesIds(context.Background(), id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative3", func(t *testing.T) {
		id, _ := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		_, err = badger.GetEdges(context.Background(), id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestBadgerCount(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(id, ids)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		n, err := badger.CountImages(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		n, err = badger.CountImagesCompressed(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 0, n)
		err = badger.UpdateCompressionStatus(context.Background(), ids.Uint64(0), ids.Uint64(1))
		assert.NoError(t, err)
		n, err = badger.CountImages(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		n, err = badger.CountImagesCompressed(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
		err = badger.UpdateCompressionStatus(context.Background(), ids.Uint64(0), ids.Uint64(2))
		assert.NoError(t, err)
		n, err = badger.CountImagesCompressed(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 2, n)
	})
	t.Run("Negative1", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(id, ids)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		err = badger.UpdateCompressionStatus(context.Background(), ids.Uint64(0), ids.Uint64(1))
		assert.NoError(t, err)
		err = badger.UpdateCompressionStatus(context.Background(), ids.Uint64(0), ids.Uint64(1))
		assert.NoError(t, err)
		n, err := badger.CountImagesCompressed(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
	})
	t.Run("Negative2", func(t *testing.T) {
		id, _ := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		_, err = badger.CountImages(context.Background(), id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative3", func(t *testing.T) {
		id, _ := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		_, err = badger.CountImagesCompressed(context.Background(), id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative4", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		err = badger.UpdateCompressionStatus(context.Background(), id(), ids.Uint64(0))
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative5", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(id, ids)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		err = badger.UpdateCompressionStatus(context.Background(), ids.Uint64(0), id())
		assert.ErrorIs(t, err, domain.ErrImageNotFound)
	})
}

func TestBadgerImage(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(id, ids)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		src, err := badger.GetImageSrc(context.Background(), ids.Uint64(0), ids.Uint64(4))
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/"+ids.Base64(0)+"/images/"+ids.Base64(4), src)
	})
	t.Run("Negative1", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		_, err = badger.GetImageSrc(context.Background(), id(), ids.Uint64(0))
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Negative2", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(id, ids)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		_, err = badger.GetImageSrc(context.Background(), ids.Uint64(0), id())
		assert.ErrorIs(t, err, domain.ErrImageNotFound)
	})
}

func TestBadgerVote(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumFullFactory(id, ids)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		err = badger.SaveVote(context.Background(), ids.Uint64(0), ids.Uint64(3), ids.Uint64(5))
		assert.NoError(t, err)
		err = badger.SaveVote(context.Background(), ids.Uint64(0), ids.Uint64(3), ids.Uint64(5))
		assert.NoError(t, err)
		edgs, err := badger.GetEdges(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 2, edgs[ids.Uint64(3)][ids.Uint64(5)])
	})
	t.Run("Negative", func(t *testing.T) {
		id, _ := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		err = badger.SaveVote(context.Background(), id(), id(), id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestBadgerSort(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumFullFactory(id, ids)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		imgs1, err := badger.GetImagesOrdered(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		img1 := model.Image{Id: ids.Uint64(4), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(4), Rating: 0.77920413}
		img2 := model.Image{Id: ids.Uint64(1), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(1), Rating: 0.48954984}
		img3 := model.Image{Id: ids.Uint64(3), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(3), Rating: 0.41218211}
		img4 := model.Image{Id: ids.Uint64(2), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(2), Rating: 0.19186324}
		img5 := model.Image{Id: ids.Uint64(5), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(5), Rating: 0.13278389}
		imgs2 := []model.Image{img1, img2, img3, img4, img5}
		assert.Equal(t, imgs2, imgs1)
	})
	t.Run("Negative", func(t *testing.T) {
		id, _ := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		_, err = badger.GetImagesOrdered(context.Background(), id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestBadgerRatings(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumFullFactory(id, ids)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		img1 := model.Image{Id: ids.Uint64(1), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(1), Rating: 0.54412788}
		img2 := model.Image{Id: ids.Uint64(2), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(2), Rating: 0.32537162}
		img3 := model.Image{Id: ids.Uint64(3), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(3), Rating: 0.43185491}
		img4 := model.Image{Id: ids.Uint64(4), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(4), Rating: 0.57356209}
		img5 := model.Image{Id: ids.Uint64(5), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(5), Rating: 0.61438023}
		imgs1 := []model.Image{img1, img2, img3, img4, img5}
		vector := map[uint64]float64{}
		vector[img1.Id] = img1.Rating
		vector[img2.Id] = img2.Rating
		vector[img3.Id] = img3.Rating
		vector[img4.Id] = img4.Rating
		vector[img5.Id] = img5.Rating
		err = badger.UpdateRatings(context.Background(), ids.Uint64(0), vector)
		assert.NoError(t, err)
		imgs2, err := badger.GetImagesOrdered(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		sort.Slice(imgs1, func(i, j int) bool { return imgs1[i].Rating > imgs1[j].Rating })
		assert.Equal(t, imgs1, imgs2)
	})
	t.Run("Negative", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		album := id()
		img1 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(1), Rating: 0.54412788}
		img2 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(2), Rating: 0.32537162}
		img3 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(3), Rating: 0.43185491}
		img4 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(4), Rating: 0.57356209}
		img5 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(5), Rating: 0.61438023}
		vector := map[uint64]float64{}
		vector[img1.Id] = img1.Rating
		vector[img2.Id] = img2.Rating
		vector[img3.Id] = img3.Rating
		vector[img4.Id] = img4.Rating
		vector[img5.Id] = img5.Rating
		err = badger.UpdateRatings(context.Background(), album, vector)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestBadgerDelete(t *testing.T) {
	t.Run("Positive1", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(id, ids)
		_, err = badger.CountImages(context.Background(), ids.Uint64(0))
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		n, err := badger.CountImages(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		albums, err := badger.AlbumsToBeDeleted(context.Background())
		assert.NoError(t, err)
		assert.Len(t, albums, 0)
		err = badger.DeleteAlbum(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		_, err = badger.CountImages(context.Background(), ids.Uint64(0))
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	t.Run("Positive2", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(id, ids)
		alb.Expires = time.Now().Add(-1 * time.Hour)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		albums, err := badger.AlbumsToBeDeleted(context.Background())
		assert.NoError(t, err)
		assert.True(t, len(albums) == 1 && albums[0].Id == alb.Id && !albums[0].Expires.IsZero())
		err = badger.DeleteAlbum(context.Background(), ids.Uint64(0))
		assert.NoError(t, err)
		_, err = badger.CountImages(context.Background(), ids.Uint64(0))
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
		t.Cleanup(func() { _ = badger.DeleteAlbum(context.Background(), ids.Uint64(0)) })
	})
	t.Run("Negative", func(t *testing.T) {
		id, ids := GenId()
		badger, err := NewBadger(DefaultBadgerConfig)
		require.NoError(t, err)
		alb := AlbumEmptyFactory(id, ids)
		err = badger.SaveAlbum(context.Background(), alb)
		assert.NoError(t, err)
		err = badger.DeleteAlbum(context.Background(), id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func TestBadgerLru(t *testing.T) {
	id, ids := GenId()
	badger, err := NewBadger(DefaultBadgerConfig)
	require.NoError(t, err)
	alb1 := AlbumEmptyFactory(id, ids)
	err = badger.SaveAlbum(context.Background(), alb1)
	assert.NoError(t, err)
	alb2 := AlbumEmptyFactory(id, ids)
	err = badger.SaveAlbum(context.Background(), alb2)
	assert.NoError(t, err)
	edgs, err := badger.GetEdges(context.Background(), ids.Uint64(0))
	assert.NoError(t, err)
	assert.Equal(t, alb1.Edges, edgs)
}
