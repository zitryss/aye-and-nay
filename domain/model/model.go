package model

import (
	"io"
)

type Album struct {
	Id     string
	Images []Image
	Edges  map[string]map[string]int
}

type Image struct {
	Id         string
	Src        string
	Token      string
	Rating     float64
	Compressed bool
}

type File struct {
	io.Reader
	Size int64
}
