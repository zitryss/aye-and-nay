package base64

import (
	"flag"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	unit        = flag.Bool("unit", false, "")
	integration = flag.Bool("int", false, "")
	ci          = flag.Bool("ci", false, "")
)

func TestFromUint64(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	tests := []struct {
		u    uint64
		want string
	}{
		{
			u:    0x0,
			want: "AAAAAAAAAAA",
		},
		{
			u:    0x1,
			want: "AQAAAAAAAAA",
		},
		{
			u:    math.MaxUint64,
			want: "__________8",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := FromUint64(tt.u)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToUint64(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	tests := []struct {
		s       string
		want    uint64
		wantErr bool
	}{
		{
			s:       "AAAAAAAAAAA",
			want:    0x0,
			wantErr: false,
		},
		{
			s:       "AQAAAAAAAAA",
			want:    0x1,
			wantErr: false,
		},
		{
			s:       "__________8",
			want:    math.MaxUint64,
			wantErr: false,
		},
		{
			s:       "00000000000",
			want:    0x4dd3344dd3344dd3,
			wantErr: false,
		},
		{
			s:       "00000000001",
			want:    0x4dd3344dd3344dd3,
			wantErr: false,
		},
		{
			s:       "",
			want:    0x0,
			wantErr: true,
		},
		{
			s:       "A",
			want:    0x0,
			wantErr: true,
		},
		{
			s:       "!@#$%^&*()_",
			want:    0x0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got, err := ToUint64(tt.s)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// go test -fuzz=FuzzFromUint64 -fuzztime 5s
func FuzzFromUint64(f *testing.F) {
	f.Add(uint64(0))
	f.Fuzz(func(t *testing.T, u1 uint64) {
		if !*unit {
			t.Skip()
		}
		b64 := FromUint64(u1)
		u2, err := ToUint64(b64)
		assert.NoError(t, err)
		assert.Equal(t, u1, u2)
	})
}

// go test -fuzz=FuzzToUint64 -fuzztime 5s
func FuzzToUint64(f *testing.F) {
	f.Add("AAAAAAAAAAA")
	f.Fuzz(func(t *testing.T, s string) {
		if !*unit {
			t.Skip()
		}
		_, err := ToUint64(s)
		if err != nil {
			t.Skip()
		}
	})
}
