package storage

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"

	minios3 "github.com/minio/minio-go/v6"

	"github.com/zitryss/aye-and-nay/internal/pool"
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

func (m *minio) Put(ctx context.Context, album string, image string, b []byte) (string, error) {
	filename := "albums/" + album + "/images/" + image
	buf := pool.GetBuffer()
	defer pool.PutBuffer(buf)
	buf.Write(b)
	_, err := m.client.PutObjectWithContext(ctx, "aye-and-nay", filename, buf, int64(buf.Len()), minios3.PutObjectOptions{})
	if err != nil {
		return "", errors.Wrap(err)
	}
	src := m.conf.prefix + "/aye-and-nay/" + filename
	return src, nil
}

func (m *minio) Get(ctx context.Context, album string, image string) ([]byte, error) {
	filename := "albums/" + album + "/images/" + image
	src, err := m.client.GetObjectWithContext(ctx, "aye-and-nay", filename, minios3.GetObjectOptions{})
	if err != nil {
		return nil, errors.Wrap(err)
	}
	dst := pool.GetBuffer()
	defer pool.PutBuffer(dst)
	_, err = io.Copy(dst, src)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return dst.Bytes(), nil
}

func (m *minio) Remove(_ context.Context, album string, image string) error {
	filename := "albums/" + album + "/images/" + image
	err := m.client.RemoveObject("aye-and-nay", filename)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
