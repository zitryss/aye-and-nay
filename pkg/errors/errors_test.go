package errors_test

import (
	"flag"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zitryss/aye-and-nay/pkg/errors"
)

var (
	unit        = flag.Bool("unit", false, "")
	integration = flag.Bool("int", false, "")
	ci          = flag.Bool("ci", false, "")
)

func TestCause(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	tests := []struct {
		give error
		want error
	}{
		{
			give: nil,
			want: nil,
		},
		{
			give: errors.New("1"),
			want: errors.New("1"),
		},
		{
			give: fmt.Errorf("2: %w", errors.New("1")),
			want: errors.New("1"),
		},
		{
			give: errors.Wrap(nil),
			want: nil,
		},
		{
			give: errors.Wrapf(nil, ""),
			want: nil,
		},
		{
			give: errors.Wrap(errors.New("1")),
			want: errors.New("1"),
		},
		{
			give: fmt.Errorf("3: %w", fmt.Errorf("2: %w", errors.New("1"))),
			want: errors.New("1"),
		},
		{
			give: fmt.Errorf("3: %w", errors.Wrap(errors.New("1"))),
			want: errors.New("1"),
		},
		{
			give: errors.Wrap(fmt.Errorf("2: %w", errors.New("1"))),
			want: errors.New("1"),
		},
		{
			give: errors.Wrap(errors.Wrap(errors.New("1"))),
			want: errors.New("1"),
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := errors.Cause(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}
