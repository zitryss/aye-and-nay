package model

import (
	"io"
)

type Album struct {
	Id     uint64
	Images []Image
	Edges  map[uint64]map[uint64]int
}

type Image struct {
	Id         uint64
	Src        string
	Token      uint64
	Rating     float64
	Compressed bool
}

type File struct {
	io.Reader
	Size int64
}
