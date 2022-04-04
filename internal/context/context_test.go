package context

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
	t.Run("Positive", func(t *testing.T) {
		ctx := WithRequestId(context.Background())
		assert.Equal(t, uint64(1), GetRequestId(ctx))
		ctx = WithRequestId(ctx)
		assert.Equal(t, uint64(2), GetRequestId(ctx))
	})
	t.Run("Negative", func(t *testing.T) {
		ctx := context.Background()
		assert.Equal(t, uint64(0), GetRequestId(ctx))
	})
}
