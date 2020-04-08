package storage

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"

	minios3 "github.com/minio/minio-go/v6"
	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

func NewMinio() (minio, error) {
	conf := newMinioConfig()
	client, err := minios3.New(conf.host+":"+conf.port, conf.accessKey, conf.secretKey, conf.secure)
	if err != nil {
		return minio{}, errors.Wrap(err)
	}
	err = retry.Do(conf.times, conf.pause, func() error {
		c := http.Client{Timeout: conf.timeout}
		resp, err := c.Get("http://" + conf.host + ":" + conf.port + "/minio/health/live")
		if err != nil {
			return errors.Wrap(err)
		}
		_, err = io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			return errors.Wrap(err)
		}
		err = resp.Body.Close()
		if err != nil {
			return errors.Wrap(err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return errors.Wrap(errors.New("no connection to minio"))
		}
		resp, err = c.Get("http://" + conf.host + ":" + conf.port + "/minio/health/ready")
		if err != nil {
			return errors.Wrap(err)
		}
		_, err = io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			return errors.Wrap(err)
		}
		err = resp.Body.Close()
		if err != nil {
			return errors.Wrap(err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return errors.Wrap(errors.New("minio is not ready"))
		}
		_, err = client.ListBuckets()
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return minio{}, errors.Wrap(err)
	}
	found, err := client.BucketExists("aye-and-nay")
	if err != nil {
		return minio{}, errors.Wrap(err)
	}
	if !found {
		err = client.MakeBucket("aye-and-nay", conf.location)
		if err != nil {
			return minio{}, errors.Wrap(err)
		}
		policy := `{"Statement":[{"Action":["s3:GetObject"],"Effect":"Allow","Principal":"*","Resource":["arn:aws:s3:::aye-and-nay/albums/*"]}],"Version":"2012-10-17"}`
		err = client.SetBucketPolicy("aye-and-nay", policy)
		if err != nil {
			return minio{}, errors.Wrap(err)
		}
	}
	return minio{conf, client}, nil
}

type minio struct {
	conf   minioConfig
	client *minios3.Client
}

func (m *minio) Upload(ctx context.Context, album string, imgs []model.Image) error {
	sem := make(chan struct{}, m.conf.connections)
	g, ctx := errgroup.WithContext(ctx)
	for i := range imgs {
		sem <- struct{}{}
		img := &imgs[i]
		g.Go(func() (e error) {
			defer func() { <-sem }()
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				err, ok := v.(error)
				if ok {
					e = errors.Wrap(err)
				} else {
					e = errors.Wrapf(model.ErrUnknown, "%v", v)
				}
			}()
			filename := "albums/" + album + "/images/" + img.Id
			buf := bytes.NewBuffer(img.B)
			img.B = nil
			_, err := m.client.PutObjectWithContext(ctx, "aye-and-nay", filename, buf, int64(buf.Len()), minios3.PutObjectOptions{})
			if err != nil {
				e = errors.Wrap(err)
				return
			}
			img.Src = m.conf.prefix + "/aye-and-nay/" + filename
			return
		})
	}
	for i := 0; i < m.conf.connections; i++ {
		sem <- struct{}{}
	}
	err := g.Wait()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
