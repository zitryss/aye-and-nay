package model

import (
	"time"
)

type Album struct {
	Id      uint64
	Images  []Image
	Edges   map[uint64]map[uint64]int
	Expires time.Time
}
