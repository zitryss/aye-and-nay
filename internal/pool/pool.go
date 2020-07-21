package pool

import (
	"bytes"
	"sync"
)

var (
	p = &sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

func GetBuffer() *bytes.Buffer {
	return p.Get().(*bytes.Buffer)
}

func PutBuffer(buf *bytes.Buffer) {
	buf.Reset()
	p.Put(buf)
}
