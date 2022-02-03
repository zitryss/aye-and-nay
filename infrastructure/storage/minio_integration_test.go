//go:build integration

package storage

import (
	"context"
	"testing"

	minioS3 "github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestMinio(t *testing.T) {
	t.Run("", func(t *testing.T) {
		minio, err := NewMinio(context.Background(), DefaultMinioConfig)
		require.NoError(t, err)
		f, err := minio.Get(context.Background(), 0x70D8, 0xD5C7)
		e := minioS3.ErrorResponse{}
		assert.ErrorAs(t, err, &e)
		assert.Nil(t, f.Reader)
		src, err := minio.Put(context.Background(), 0x70D8, 0xD5C7, Png())
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/2HAAAAAAAAA/images/x9UAAAAAAAA", src)
		f, err = minio.Get(context.Background(), 0x70D8, 0xD5C7)
		assert.NoError(t, err)
		AssertEqualFile(t, f, Png())
		err = minio.Remove(context.Background(), 0x70D8, 0xD5C7)
		assert.NoError(t, err)
		f, err = minio.Get(context.Background(), 0x70D8, 0xD5C7)
		e = minioS3.ErrorResponse{}
		assert.ErrorAs(t, err, &e)
		assert.Nil(t, f.Reader)
	})
	t.Run("", func(t *testing.T) {
		minio, err := NewMinio(context.Background(), DefaultMinioConfig)
		require.NoError(t, err)
		src, err := minio.Put(context.Background(), 0x872D, 0x882D, Png())
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/LYcAAAAAAAA/images/LYgAAAAAAAA", src)
		f, err := minio.Get(context.Background(), 0x872D, 0x882D)
		assert.NoError(t, err)
		AssertEqualFile(t, f, Png())
		err = minio.Remove(context.Background(), 0x872D, 0x882D)
		assert.NoError(t, err)
		src, err = minio.Put(context.Background(), 0x872D, 0x882D, Png())
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/LYcAAAAAAAA/images/LYgAAAAAAAA", src)
	})
}

func TestMinioHealth(t *testing.T) {
	minio, err := NewMinio(context.Background(), DefaultMinioConfig)
	require.NoError(t, err)
	_, err = minio.Health(context.Background())
	assert.NoError(t, err)
}
