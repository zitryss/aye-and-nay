// +build integration

package storage

import (
	"context"
	"os"
	"testing"

	"github.com/zitryss/aye-and-nay/domain/model"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	"github.com/zitryss/aye-and-nay/internal/dockertest"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/env"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func TestMain(m *testing.M) {
	_, err := env.Lookup("CONTINUOUS_INTEGRATION")
	if err != nil {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.Lcritical)
		docker := dockertest.New()
		docker.RunMinio()
		code := m.Run()
		docker.Purge()
		os.Exit(code)
	}
	code := m.Run()
	os.Exit(code)
}

func TestMinioUpload(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		minio, err := NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		img1 := model.Image{Id: "bx3TWdDEQrF294dx", B: Png()}
		img2 := model.Image{Id: "JtnKxUarcWeJ342R", B: Png()}
		imgs := []model.Image{img1, img2}
		err = minio.Upload(context.Background(), "42QbW4V2", imgs)
		if err != nil {
			t.Error(err)
		}
		if imgs[0].B != nil {
			t.Error("imgs[0].B != nil")
		}
		if imgs[1].B != nil {
			t.Error("imgs[1].B != nil")
		}
		if imgs[0].Src != "/aye-and-nay/albums/42QbW4V2/images/bx3TWdDEQrF294dx" {
			t.Error("imgs[0].Src")
		}
		if imgs[1].Src != "/aye-and-nay/albums/42QbW4V2/images/JtnKxUarcWeJ342R" {
			t.Error("imgs[1].Src")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		minio, err := NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		img1 := model.Image{Id: "GXmZcfJZz4mRDq5R", B: nil}
		img2 := model.Image{Id: "VTq25ujPhDYsxbJ2", B: nil}
		imgs := []model.Image{img1, img2}
		err = minio.Upload(context.Background(), "bSTApWdG", imgs)
		if err != nil {
			t.Error(err)
		}
		if imgs[0].B != nil {
			t.Error("imgs[0].B != nil")
		}
		if imgs[1].B != nil {
			t.Error("imgs[1].B != nil")
		}
		if imgs[0].Src != "/aye-and-nay/albums/bSTApWdG/images/GXmZcfJZz4mRDq5R" {
			t.Error("imgs[0].Src")
		}
		if imgs[1].Src != "/aye-and-nay/albums/bSTApWdG/images/VTq25ujPhDYsxbJ2" {
			t.Error("imgs[1].Src")
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		minio, err := NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		img1 := model.Image{Id: "tkNzABZkEw2PdVHv", B: Png()}
		imgs := []model.Image{img1}
		err = minio.Upload(context.Background(), "fhR5PQN6", imgs)
		if err != nil {
			t.Error(err)
		}
		if imgs[0].B != nil {
			t.Error("imgs[0].B != nil")
		}
		if imgs[0].Src != "/aye-and-nay/albums/fhR5PQN6/images/tkNzABZkEw2PdVHv" {
			t.Error("imgs[0].Src")
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		minio, err := NewMinio()
		if err != nil {
			t.Fatal(err)
		}
		imgs := []model.Image(nil)
		err = minio.Upload(context.Background(), "3dUtqgkY", imgs)
		if err != nil {
			t.Error(err)
		}
	})
}
