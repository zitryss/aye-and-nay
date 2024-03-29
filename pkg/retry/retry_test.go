package retry_test

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

var (
	unit        = flag.Bool("unit", false, "")
	integration = flag.Bool("int", false, "")
	ci          = flag.Bool("ci", false, "")
)

func TestDo1(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	type give struct {
		times int
		pause time.Duration
		busy  time.Duration
	}
	type want struct {
		err   error
		calls int
	}
	tests := []struct {
		give
		want
	}{
		{
			give{
				times: 0,
			},
			want{
				err:   nil,
				calls: 1,
			},
		},
		{
			give{
				times: 1,
			},
			want{
				err:   nil,
				calls: 1,
			},
		},
		{
			give{
				times: 2,
			},
			want{
				err:   nil,
				calls: 1,
			},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			c := 0
			err := retry.Do(tt.times, tt.pause, func() error {
				c++
				time.Sleep(tt.busy)
				return nil
			})
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.calls, c)
		})
	}
}

func TestDo2(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	type give struct {
		times int
		pause time.Duration
		busy  time.Duration
	}
	type want struct {
		err   error
		calls int
	}
	tests := []struct {
		give
		want
	}{
		{
			give{
				times: 0,
			},
			want{
				err:   errors.New("no luck"),
				calls: 1,
			},
		},
		{
			give{
				times: 1,
			},
			want{
				err:   errors.New("no luck"),
				calls: 2,
			},
		},
		{
			give{
				times: 2,
			},
			want{
				err:   errors.New("no luck"),
				calls: 3,
			},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			c := 0
			err := retry.Do(tt.times, tt.pause, func() error {
				c++
				time.Sleep(tt.busy)
				return errors.New("no luck")
			})
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.calls, c)
		})
	}
}

func TestDo3(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	type give struct {
		times int
		pause time.Duration
		busy  time.Duration
	}
	type want struct {
		err   error
		calls int
	}
	tests := []struct {
		give
		want
	}{
		{
			give{
				times: 0,
			},
			want{
				err:   errors.New("no luck"),
				calls: 1,
			},
		},
		{
			give{
				times: 1,
			},
			want{
				err:   nil,
				calls: 2,
			},
		},
		{
			give{
				times: 2,
			},
			want{
				err:   nil,
				calls: 2,
			},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			c := 0
			err := retry.Do(tt.times, tt.pause, func() error {
				c++
				time.Sleep(tt.busy)
				if c == 1 {
					return errors.New("no luck")
				}
				return nil
			})
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.calls, c)
		})
	}
}
