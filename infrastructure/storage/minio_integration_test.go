//go:build integration

package storage

import (
	"context"
	"testing"

	minioS3 "github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/zitryss/aye-and-nay/internal/generator"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestMinio(t *testing.T) {
	t.Run("", func(t *testing.T) {
		id, ids := GenId()
		album := id()
		image := id()
		minio, err := NewMinio(context.Background(), DefaultMinioConfig)
		require.NoError(t, err)
		f, err := minio.Get(context.Background(), album, image)
		e := minioS3.ErrorResponse{}
		assert.ErrorAs(t, err, &e)
		assert.Nil(t, f.Reader)
		src, err := minio.Put(context.Background(), album, image, Png())
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/"+ids.Base64(0)+"/images/"+ids.Base64(1), src)
		f, err = minio.Get(context.Background(), album, image)
		assert.NoError(t, err)
		AssertEqualFile(t, f, Png())
		err = minio.Remove(context.Background(), album, image)
		assert.NoError(t, err)
		f, err = minio.Get(context.Background(), album, image)
		e = minioS3.ErrorResponse{}
		assert.ErrorAs(t, err, &e)
		assert.Nil(t, f.Reader)
	})
	t.Run("", func(t *testing.T) {
		id, ids := GenId()
		album := id()
		image := id()
		minio, err := NewMinio(context.Background(), DefaultMinioConfig)
		require.NoError(t, err)
		src, err := minio.Put(context.Background(), album, image, Png())
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/"+ids.Base64(0)+"/images/"+ids.Base64(1), src)
		f, err := minio.Get(context.Background(), album, image)
		assert.NoError(t, err)
		AssertEqualFile(t, f, Png())
		err = minio.Remove(context.Background(), album, image)
		assert.NoError(t, err)
		src, err = minio.Put(context.Background(), album, image, Png())
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/"+ids.Base64(0)+"/images/"+ids.Base64(1), src)
	})
}

func TestMinioHealth(t *testing.T) {
	minio, err := NewMinio(context.Background(), DefaultMinioConfig)
	require.NoError(t, err)
	_, err = minio.Health(context.Background())
	assert.NoError(t, err)
}
