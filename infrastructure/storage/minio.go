package storage

import (
	"context"
	"io"
	"net/http"

	minioS3 "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/base64"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/pool"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

var (
	_ domain.Storager = (*Minio)(nil)
)

func NewMinio(ctx context.Context, conf MinioConfig) (*Minio, error) {
	client, err := minioS3.New(conf.Host+":"+conf.Port, &minioS3.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, conf.Token),
		Secure: conf.Secure,
	})
	if err != nil {
		return &Minio{}, errors.Wrap(err)
	}
	m := &Minio{conf, client}
	ctx, cancel := context.WithTimeout(ctx, conf.Timeout)
	defer cancel()
	err = retry.Do(conf.RetryTimes, conf.RetryPause, func() error {
		_, err := m.Health(ctx)
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
		err = client.MakeBucket(ctx, "aye-and-nay", minioS3.MakeBucketOptions{Region: conf.Location})
		if err != nil {
			return &Minio{}, errors.Wrap(err)
		}
		policy := `{"Statement":[{"Action":["s3:GetObject"],"Effect":"Allow","Principal":"*","Resource":["arn:aws:s3:::aye-and-nay/albums/*"]}],"Version":"2012-10-17"}`
		err = client.SetBucketPolicy(ctx, "aye-and-nay", policy)
		if err != nil {
			return &Minio{}, errors.Wrap(err)
		}
	}
	return m, nil
}

type Minio struct {
	conf   MinioConfig
	client *minioS3.Client
}

func (m *Minio) Put(ctx context.Context, album uint64, image uint64, f model.File) (string, error) {
	defer f.Close()
	albumB64 := base64.FromUint64(album)
	imageB64 := base64.FromUint64(image)
	filename := "albums/" + albumB64 + "/images/" + imageB64
	_, err := m.client.PutObject(ctx, "aye-and-nay", filename, f.Reader, f.Size, minioS3.PutObjectOptions{})
	if err != nil {
		return "", errors.Wrap(err)
	}
	src := m.conf.Prefix + "/aye-and-nay/" + filename
	return src, nil
}

func (m *Minio) Get(ctx context.Context, album uint64, image uint64) (model.File, error) {
	albumB64 := base64.FromUint64(album)
	imageB64 := base64.FromUint64(image)
	filename := "albums/" + albumB64 + "/images/" + imageB64
	obj, err := m.client.GetObject(ctx, "aye-and-nay", filename, minioS3.GetObjectOptions{})
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	defer obj.Close()
	info, err := obj.Stat()
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	buf := pool.GetBufferN(info.Size)
	n, err := io.Copy(buf, obj)
	if err != nil {
		return model.File{}, errors.Wrap(err)
	}
	closeFn := func() error {
		pool.PutBuffer(buf)
		return nil
	}
	return model.NewFile(buf, closeFn, n), nil
}

func (m *Minio) Remove(ctx context.Context, album uint64, image uint64) error {
	albumB64 := base64.FromUint64(album)
	imageB64 := base64.FromUint64(image)
	filename := "albums/" + albumB64 + "/images/" + imageB64
	err := m.client.RemoveObject(ctx, "aye-and-nay", filename, minioS3.RemoveObjectOptions{})
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (m *Minio) Health(ctx context.Context) (bool, error) {
	url := "http://" + m.conf.Host + ":" + m.conf.Port + "/minio/health/live"
	body := io.Reader(nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, body)
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthStorage, "%s", err)
	}
	c := http.Client{Timeout: m.conf.Timeout}
	resp, err := c.Do(req)
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthStorage, "%s", err)
	}
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return false, errors.Wrapf(domain.ErrBadHealthStorage, "%s", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthStorage, "%s", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return false, errors.Wrapf(domain.ErrBadHealthStorage, "%s", "no connection to minio")
	}
	url = "http://" + m.conf.Host + ":" + m.conf.Port + "/minio/health/ready"
	body = io.Reader(nil)
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, body)
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthStorage, "%s", err)
	}
	c = http.Client{Timeout: m.conf.Timeout}
	resp, err = c.Do(req)
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthStorage, "%s", err)
	}
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return false, errors.Wrapf(domain.ErrBadHealthStorage, "%s", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthStorage, "%s", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return false, errors.Wrapf(domain.ErrBadHealthStorage, "%s", "minio is not ready")
	}
	_, err = m.client.ListBuckets(ctx)
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthStorage, "%s", err)
	}
	return true, nil
}
