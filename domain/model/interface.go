package model

import (
	"context"
	"time"
)

type Servicer interface {
	Album(ctx context.Context, ff []File, dur time.Duration) (string, error)
	Pair(ctx context.Context, album string) (Image, Image, error)
	Vote(ctx context.Context, album string, tokenFrom string, tokenTo string) error
	Top(ctx context.Context, album string) ([]Image, error)
	Progress(ctx context.Context, album string) (float64, error)
}

type Compresser interface {
	Compress(ctx context.Context, f File) (File, error)
}

type Storager interface {
	Put(ctx context.Context, album string, image string, f File) (string, error)
	Get(ctx context.Context, album string, image string) (File, error)
	Remove(ctx context.Context, album string, image string) error
}

type Persister interface {
	SaveAlbum(ctx context.Context, alb Album) error
	CountImages(ctx context.Context, album string) (int, error)
	CountImagesCompressed(ctx context.Context, album string) (int, error)
	UpdateCompressionStatus(ctx context.Context, album string, image string) error
	GetImage(ctx context.Context, album string, image string) (Image, error)
	GetImages(ctx context.Context, album string) ([]string, error)
	SaveVote(ctx context.Context, album string, imageFrom string, imageTo string) error
	GetEdges(ctx context.Context, album string) (map[string]map[string]int, error)
	UpdateRatings(ctx context.Context, album string, vector map[string]float64) error
	GetImagesOrdered(ctx context.Context, album string) ([]Image, error)
	DeleteAlbum(ctx context.Context, album string) error
}

type Temper interface {
	Queuer
	PQueuer
	Stacker
	Tokener
}

type Queuer interface {
	Add(ctx context.Context, queue string, album string) error
	Poll(ctx context.Context, queue string) (string, error)
	Size(ctx context.Context, queue string) (int, error)
}

type PQueuer interface {
	PAdd(ctx context.Context, pqueue string, album string, expires time.Time) error
	PPoll(ctx context.Context, pqueue string) (string, time.Time, error)
	PSize(ctx context.Context, pqueue string) (int, error)
}

type Stacker interface {
	Push(ctx context.Context, album string, pairs [][2]string) error
	Pop(ctx context.Context, album string) (string, string, error)
}

type Tokener interface {
	Set(ctx context.Context, album string, token string, image string) error
	Get(ctx context.Context, album string, token string) (string, error)
}
