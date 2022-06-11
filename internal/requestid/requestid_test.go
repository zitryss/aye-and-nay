package requestid

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	unit        = flag.Bool("unit", false, "")
	integration = flag.Bool("int", false, "")
	ci          = flag.Bool("ci", false, "")
)

func TestContext(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	t.Run("Positive", func(t *testing.T) {
		ctx := Set(context.Background())
		id1 := Get(ctx)
		assert.NotZero(t, id1)
		assert.Positive(t, id1)
		assert.Greater(t, id1, uint64(0))
		ctx = Set(ctx)
		id2 := Get(ctx)
		assert.NotZero(t, id2)
		assert.Positive(t, id2)
		assert.Greater(t, id2, id1)
	})
	t.Run("Negative", func(t *testing.T) {
		ctx := context.Background()
		assert.Equal(t, uint64(0), Get(ctx))
	})
}
