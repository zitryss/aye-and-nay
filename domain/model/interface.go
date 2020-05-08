package model

import (
	"context"
)

type Servicer interface {
	Album(ctx context.Context, files [][]byte) (string, error)
	Pair(ctx context.Context, album string) (Image, Image, error)
	Vote(ctx context.Context, album string, tokenFrom string, tokenTo string) error
	Top(ctx context.Context, album string) ([]Image, error)
	Exists(ctx context.Context, album string) (bool, error)
}

type Compresser interface {
	Compress(ctx context.Context, imgs []Image) error
}

type Storager interface {
	Put(ctx context.Context, album string, image string, b []byte) (string, error)
	Get(ctx context.Context, album string, image string) ([]byte, error)
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
	CheckAlbum(ctx context.Context, album string) (bool, error)
}

type Temper interface {
	Queuer
	Stacker
	Tokener
}

type Queuer interface {
	Add(ctx context.Context, queue string, album string) error
	Poll(ctx context.Context, queue string) (string, error)
	Size(ctx context.Context, queue string) (int, error)
}

type Stacker interface {
	Push(ctx context.Context, album string, pairs [][2]string) error
	Pop(ctx context.Context, album string) (string, string, error)
}

type Tokener interface {
	Set(ctx context.Context, album string, token string, image string) error
	Get(ctx context.Context, album string, token string) (string, error)
}
