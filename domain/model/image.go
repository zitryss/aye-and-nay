package model

type Image struct {
	Id         uint64
	Src        string
	Token      uint64
	Rating     float64
	Compressed bool
}
