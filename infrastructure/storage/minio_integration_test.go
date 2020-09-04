// +build integration

package storage

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	minios3 "github.com/minio/minio-go/v7"

	_ "github.com/zitryss/aye-and-nay/internal/config"
	"github.com/zitryss/aye-and-nay/internal/dockertest"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/env"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func TestMain(m *testing.M) {
	_, err := env.Lookup("CONTINUOUS_INTEGRATION")
	if err != nil {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.Lcritical)
		docker := dockertest.New()
		docker.RunMinio()
		log.SetOutput(ioutil.Discard)
		code := m.Run()
		docker.Purge()
		os.Exit(code)
	}
	code := m.Run()
	os.Exit(code)
}

func TestMinio(t *testing.T) {
	t.Run("", func(t *testing.T) {
		minio, err := NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		f, err := minio.Get(context.Background(), "pXFAGSZpY844QjvY", "qmgc5mNJtxUtF8WU")
		e := (*minios3.ErrorResponse)(nil)
		if errors.As(err, &e) {
			t.Error(err)
		}
		if f.Reader != nil {
			t.Error("f.Reader != nil")
		}
		src, err := minio.Put(context.Background(), "pXFAGSZpY844QjvY", "qmgc5mNJtxUtF8WU", Png())
		if err != nil {
			t.Error(err)
		}
		if src != "/aye-and-nay/albums/pXFAGSZpY844QjvY/images/qmgc5mNJtxUtF8WU" {
			t.Error("src != \"/aye-and-nay/albums/pXFAGSZpY844QjvY/images/qmgc5mNJtxUtF8WU\"")
		}
		f, err = minio.Get(context.Background(), "pXFAGSZpY844QjvY", "qmgc5mNJtxUtF8WU")
		if err != nil {
			t.Error(err)
		}
		if !EqualFile(f, Png()) {
			t.Error("!EqualFile(f, Png())")
		}
		err = minio.Remove(context.Background(), "pXFAGSZpY844QjvY", "qmgc5mNJtxUtF8WU")
		if err != nil {
			t.Error(err)
		}
		f, err = minio.Get(context.Background(), "pXFAGSZpY844QjvY", "qmgc5mNJtxUtF8WU")
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
		src, err := minio.Put(context.Background(), "DRkhsmc3WAsXkSFm", "aYWVaKdC8fg2e65D", Png())
		if err != nil {
			t.Error(err)
		}
		if src != "/aye-and-nay/albums/DRkhsmc3WAsXkSFm/images/aYWVaKdC8fg2e65D" {
			t.Error("src != \"/aye-and-nay/albums/DRkhsmc3WAsXkSFm/images/aYWVaKdC8fg2e65D\"")
		}
		f, err := minio.Get(context.Background(), "DRkhsmc3WAsXkSFm", "aYWVaKdC8fg2e65D")
		if err != nil {
			t.Error(err)
		}
		if !EqualFile(f, Png()) {
			t.Error("!EqualFile(f, Png())")
		}
		err = minio.Remove(context.Background(), "DRkhsmc3WAsXkSFm", "aYWVaKdC8fg2e65D")
		if err != nil {
			t.Error(err)
		}
		src, err = minio.Put(context.Background(), "DRkhsmc3WAsXkSFm", "aYWVaKdC8fg2e65D", Png())
		if err != nil {
			t.Error(err)
		}
		if src != "/aye-and-nay/albums/DRkhsmc3WAsXkSFm/images/aYWVaKdC8fg2e65D" {
			t.Error("src != \"/aye-and-nay/albums/DRkhsmc3WAsXkSFm/images/aYWVaKdC8fg2e65D\"")
		}
	})
}
