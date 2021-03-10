package storage

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"os"

	minios3 "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/internal/pool"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

func NewMinio() (*Minio, error) {
	conf := newMinioConfig()
	client, err := minios3.New(conf.host+":"+conf.port, &minios3.Options{
		Creds:  credentials.NewStaticV4(conf.accessKey, conf.secretKey, conf.token),
		Secure: conf.secure,
	})
	if err != nil {
		return &Minio{}, errors.Wrap(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), conf.timeout)
	defer cancel()
	err = retry.Do(conf.times, conf.pause, func() error {
		c := http.Client{}
		url := "http://" + conf.host + ":" + conf.port + "/minio/health/live"
		body := io.Reader(nil)
		req, err := http.NewRequestWithContext(ctx, "GET", url, body)
		if err != nil {
			return errors.Wrap(err)
		}
		resp, err := c.Do(req)
		if err != nil {
			return errors.Wrap(err)
		}
		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return errors.Wrap(err)
		}
		err = resp.Body.Close()
		if err != nil {
			return errors.Wrap(err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return errors.Wrap(errors.New("no connection to minio"))
		}
		c = http.Client{}
		url = "http://" + conf.host + ":" + conf.port + "/minio/health/ready"
		body = io.Reader(nil)
		req, err = http.NewRequestWithContext(ctx, "GET", url, body)
		if err != nil {
			return errors.Wrap(err)
		}
		resp, err = c.Do(req)
		if err != nil {
			return errors.Wrap(err)
		}
		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return errors.Wrap(err)
		}
		err = resp.Body.Close()
		if err != nil {
			return errors.Wrap(err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return errors.Wrap(errors.New("minio is not ready"))
		}
		_, err = client.ListBuckets(ctx)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return &Minio{}, errors.Wrap(err)
	}
	found, err := client.BucketExists(ctx, "aye-and-nay")
	if err != nil {
		return &Minio{}, errors.Wrap(err)
	}
	if !found {
		err = client.MakeBucket(ctx, "aye-and-nay", minios3.MakeBucketOptions{Region: conf.location})
		if err != nil {
			return &Minio{}, errors.Wrap(err)
		}
		policy := `{"Statement":[{"Action":["s3:GetObject"],"Effect":"Allow","Principal":"*","Resource":["arn:aws:s3:::aye-and-nay/albums/*"]}],"Version":"2012-10-17"}`
		err = client.SetBucketPolicy(ctx, "aye-and-nay", policy)
		if err != nil {
			return &Minio{}, errors.Wrap(err)
		}
	}
	return &Minio{conf, client}, nil
}

type Minio struct {
	conf   minioConfig
	client *minios3.Client
}

func (m *Minio) Put(ctx context.Context, album uint64, image uint64, f model.File) (string, error) {
	defer func() {
		switch v := f.Reader.(type) {
		case *os.File:
			_ = v.Close()
			_ = os.Remove(v.Name())
		case *bytes.Buffer:
			pool.PutBuffer(v)
		}
	}()
	filename := "albums/" + album + "/images/" + image
	buf := bufio.NewReader(f)
	_, err := m.client.PutObject(ctx, "aye-and-nay", filename, buf, f.Size, minios3.PutObjectOptions{})
	if err != nil {
		return "", errors.Wrap(err)
	}
	src := m.conf.prefix + "/aye-and-nay/" + filename
	return src, nil
}

func (m *Minio) Get(ctx context.Context, album uint64, image uint64) (model.File, error) {
	filename := "albums/" + album + "/images/" + image
	obj, err := m.client.GetObject(ctx, "aye-and-nay", filename, minios3.GetObjectOptions{})
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	buf := pool.GetBuffer()
	n, err := io.Copy(buf, obj)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	return model.File{Reader: buf, Size: n}, nil
}

func (m *Minio) Remove(ctx context.Context, album uint64, image uint64) error {
	filename := "albums/" + album + "/images/" + image
	err := m.client.RemoveObject(ctx, "aye-and-nay", filename, minios3.RemoveObjectOptions{})
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
