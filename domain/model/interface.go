package model

import (
	"context"
	"time"
)

type Servicer interface {
	Album(ctx context.Context, ff []File, dur time.Duration) (uint64, error)
	Pair(ctx context.Context, album uint64) (Image, Image, error)
	Vote(ctx context.Context, album uint64, tokenFrom uint64, tokenTo uint64) error
	Top(ctx context.Context, album uint64) ([]Image, error)
	Progress(ctx context.Context, album uint64) (float64, error)
}

type Compresser interface {
	Compress(ctx context.Context, f File) (File, error)
}

type Storager interface {
	Put(ctx context.Context, album uint64, image uint64, f File) (string, error)
	Get(ctx context.Context, album uint64, image uint64) (File, error)
	Remove(ctx context.Context, album uint64, image uint64) error
}

type Databaser interface {
	SaveAlbum(ctx context.Context, alb Album) error
	CountImages(ctx context.Context, album uint64) (int, error)
	CountImagesCompressed(ctx context.Context, album uint64) (int, error)
	UpdateCompressionStatus(ctx context.Context, album uint64, image uint64) error
	GetImageSrc(ctx context.Context, album uint64, image uint64) (string, error)
	GetImagesIds(ctx context.Context, album uint64) ([]uint64, error)
	SaveVote(ctx context.Context, album uint64, imageFrom uint64, imageTo uint64) error
	GetEdges(ctx context.Context, album uint64) (map[uint64]map[uint64]int, error)
	UpdateRatings(ctx context.Context, album uint64, vector map[uint64]float64) error
	GetImagesOrdered(ctx context.Context, album uint64) ([]Image, error)
	DeleteAlbum(ctx context.Context, album uint64) error
}

type Cacher interface {
	Queuer
	PQueuer
	Stacker
	Tokener
}

type Queuer interface {
	Add(ctx context.Context, queue uint64, album uint64) error
	Poll(ctx context.Context, queue uint64) (uint64, error)
	Size(ctx context.Context, queue uint64) (int, error)
}

type PQueuer interface {
	PAdd(ctx context.Context, pqueue uint64, album uint64, expires time.Time) error
	PPoll(ctx context.Context, pqueue uint64) (uint64, time.Time, error)
	PSize(ctx context.Context, pqueue uint64) (int, error)
}

type Stacker interface {
	Push(ctx context.Context, album uint64, pairs [][2]uint64) error
	Pop(ctx context.Context, album uint64) (uint64, uint64, error)
}

type Tokener interface {
	Set(ctx context.Context, album uint64, token uint64, image uint64) error
	Get(ctx context.Context, album uint64, token uint64) (uint64, error)
}
