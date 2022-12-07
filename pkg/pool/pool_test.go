package pool

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	unit        = flag.Bool("unit", false, "")
	integration = flag.Bool("int", false, "")
	ci          = flag.Bool("ci", false, "")
)

func TestBuffer(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	buf1 := GetBuffer()
	assert.NotNil(t, buf1)
	PutBuffer(buf1)
	buf2 := GetBufferN(100)
	assert.NotNil(t, buf2)
	assert.GreaterOrEqual(t, buf2.Cap(), 100)
	PutBuffer(buf2)
}
