//go:build unit

package errors_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func TestCause(t *testing.T) {
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cause(%v) = %v, want %v", tt.give, got, tt.want)
			}
		})
	}
}
