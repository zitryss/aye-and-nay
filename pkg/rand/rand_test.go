package rand

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

func TestId(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	got, err := Id()
	assert.NoError(t, err)
	assert.Positive(t, got)
}
