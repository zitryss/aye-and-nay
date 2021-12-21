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

func GetBufferN(n int64) *bytes.Buffer {
	if n <= 0 {
		return GetBuffer()
	}
	buf := GetBuffer()
	delta := buf.Cap() - int(n)
	if delta < 0 {
		buf.Grow(-delta)
	}
	return buf
}

func PutBuffer(buf *bytes.Buffer) {
	buf.Reset()
	p.Put(buf)
}
