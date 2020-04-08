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
	Upload(ctx context.Context, album string, imgs []Image) error
}

type Persister interface {
	SaveAlbum(ctx context.Context, alb Album) error
	GetImage(ctx context.Context, album string, image string) (Image, error)
	GetImages(ctx context.Context, album string) ([]string, error)
	SaveVote(ctx context.Context, album string, imageFrom string, imageTo string) error
	GetEdges(ctx context.Context, album string) (map[string]map[string]int, error)
	UpdateRatings(ctx context.Context, album string, vector map[string]float64) error
	GetImagesOrdered(ctx context.Context, album string) ([]Image, error)
	CheckAlbum(ctx context.Context, album string) (bool, error)
}

type Cacher interface {
	PopPair(ctx context.Context, album string) (string, string, error)
	PushPair(ctx context.Context, album string, pairs [][2]string) error
	GetImageId(ctx context.Context, album string, token string) (string, error)
	SetToken(ctx context.Context, album string, token string, image string) error
}
