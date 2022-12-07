package debug

import (
	"errors"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	unit        = flag.Bool("unit", false, "")
	integration = flag.Bool("int", false, "")
	ci          = flag.Bool("ci", false, "")
)

func TestAssert(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	assert.NotPanics(t, func() {
		Assert(true)
	})
	assert.Panics(t, func() {
		Assert(false)
	})
}

func TestCheck(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	assert.NotPanics(t, func() {
		Check(nil)
	})
	assert.Panics(t, func() {
		Check(errors.New(""))
	})
}
