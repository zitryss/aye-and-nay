//go:build integration

package storage

import (
	"context"
	"testing"

	minios3 "github.com/minio/minio-go/v7"

	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func TestMinio(t *testing.T) {
	t.Run("", func(t *testing.T) {
		minio, err := NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		f, err := minio.Get(context.Background(), 0x70D8, 0xD5C7)
		e := (*minios3.ErrorResponse)(nil)
		if errors.As(err, &e) {
			t.Error(err)
		}
		if f.Reader != nil {
			t.Error("f.Reader != nil")
		}
		src, err := minio.Put(context.Background(), 0x70D8, 0xD5C7, Png())
		if err != nil {
			t.Error(err)
		}
		if src != "/aye-and-nay/albums/2HAAAAAAAAA/images/x9UAAAAAAAA" {
			t.Error("src != \"/aye-and-nay/albums/2HAAAAAAAAA/images/x9UAAAAAAAA\"")
		}
		f, err = minio.Get(context.Background(), 0x70D8, 0xD5C7)
		if err != nil {
			t.Error(err)
		}
		if !EqualFile(f, Png()) {
			t.Error("!EqualFile(f, Png())")
		}
		err = minio.Remove(context.Background(), 0x70D8, 0xD5C7)
		if err != nil {
			t.Error(err)
		}
		f, err = minio.Get(context.Background(), 0x70D8, 0xD5C7)
		e = (*minios3.ErrorResponse)(nil)
		if errors.As(err, &e) {
			t.Error(err)
		}
		if f.Reader != nil {
			t.Error("f.Reader != nil")
		}
	})
	t.Run("", func(t *testing.T) {
		minio, err := NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		src, err := minio.Put(context.Background(), 0x872D, 0x882D, Png())
		if err != nil {
			t.Error(err)
		}
		if src != "/aye-and-nay/albums/LYcAAAAAAAA/images/LYgAAAAAAAA" {
			t.Error("src != \"/aye-and-nay/albums/LYcAAAAAAAA/images/LYgAAAAAAAA\"")
		}
		f, err := minio.Get(context.Background(), 0x872D, 0x882D)
		if err != nil {
			t.Error(err)
		}
		if !EqualFile(f, Png()) {
			t.Error("!EqualFile(f, Png())")
		}
		err = minio.Remove(context.Background(), 0x872D, 0x882D)
		if err != nil {
			t.Error(err)
		}
		src, err = minio.Put(context.Background(), 0x872D, 0x882D, Png())
		if err != nil {
			t.Error(err)
		}
		if src != "/aye-and-nay/albums/LYcAAAAAAAA/images/LYgAAAAAAAA" {
			t.Error("src != \"/aye-and-nay/albums/LYcAAAAAAAA/images/LYgAAAAAAAA\"")
		}
	})
}
