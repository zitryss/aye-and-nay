package gctuner

import (
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestReadCgroupMemTotal(t *testing.T) {
	tests := []struct {
		f       io.Reader
		want    float64
		wantErr bool
	}{
		{
			f:       strings.NewReader(""),
			want:    0,
			wantErr: true,
		},
		{
			f:       strings.NewReader("0"),
			want:    0,
			wantErr: false,
		},
		{
			f:       strings.NewReader("1"),
			want:    1,
			wantErr: false,
		},
		{
			f:       strings.NewReader("\r1\n"),
			want:    1,
			wantErr: false,
		},
		{
			f:       strings.NewReader("-1"),
			want:    0,
			wantErr: false,
		},
		{
			f:       strings.NewReader("9223372036854775807"),
			want:    9223372036854775807,
			wantErr: false,
		},
		{
			f:       strings.NewReader("-9223372036854775808"),
			want:    0,
			wantErr: false,
		},
		{
			f:       strings.NewReader("max"),
			want:    0,
			wantErr: false,
		},
		{
			f:       strings.NewReader("abc"),
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got, err := readCgroupMemTotal(tt.f)
			if (err != nil) != tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemTotal(t *testing.T) {
	tests := []struct {
		total            float64
		cgroupMemTotalV1 []byte
		cgroupMemTotalV2 []byte
		want             float64
	}{
		{
			total:            1073741824,
			cgroupMemTotalV1: []byte(""),
			cgroupMemTotalV2: []byte("-1"),
			want:             1073741824,
		},
		{
			total:            1073741824,
			cgroupMemTotalV1: []byte("$%^&"),
			cgroupMemTotalV2: []byte("0"),
			want:             1073741824,
		},
		{
			total:            1073741824,
			cgroupMemTotalV1: []byte("max"),
			cgroupMemTotalV2: []byte("$%^&"),
			want:             1073741824,
		},
		{
			total:            0,
			cgroupMemTotalV1: []byte("943718400\n"),
			cgroupMemTotalV2: nil,
			want:             943718400,
		},
		{
			total:            0,
			cgroupMemTotalV1: nil,
			cgroupMemTotalV2: []byte("734003200\n"),
			want:             734003200,
		},

		{
			total:            1073741824,
			cgroupMemTotalV1: []byte("943718400\n"),
			cgroupMemTotalV2: nil,
			want:             943718400,
		},
		{
			total:            943718400,
			cgroupMemTotalV1: []byte("1073741824\n"),
			cgroupMemTotalV2: nil,
			want:             943718400,
		},
		{
			total:            1073741824,
			cgroupMemTotalV1: nil,
			cgroupMemTotalV2: []byte("943718400\n"),
			want:             943718400,
		},
		{
			total:            943718400,
			cgroupMemTotalV1: nil,
			cgroupMemTotalV2: []byte("1073741824\n"),
			want:             943718400,
		},
		{
			total:            0,
			cgroupMemTotalV1: []byte("943718400\n"),
			cgroupMemTotalV2: []byte("1073741824\n"),
			want:             943718400,
		},
		{
			total:            0,
			cgroupMemTotalV1: []byte("1073741824\n"),
			cgroupMemTotalV2: []byte("943718400\n"),
			want:             943718400,
		},
		{
			total:            734003200,
			cgroupMemTotalV1: []byte("1073741824\n"),
			cgroupMemTotalV2: []byte("943718400\n"),
			want:             734003200,
		},
		{
			total:            943718400,
			cgroupMemTotalV1: []byte("1073741824\n"),
			cgroupMemTotalV2: []byte("734003200\n"),
			want:             734003200,
		},
		{
			total:            1073741824,
			cgroupMemTotalV1: []byte("734003200\n"),
			cgroupMemTotalV2: []byte("943718400\n"),
			want:             734003200,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			appFs = afero.NewMemMapFs()
			if tt.cgroupMemTotalV1 != nil {
				err := afero.WriteFile(appFs, cgroupMemTotalPathV1, tt.cgroupMemTotalV1, 0644)
				assert.NoError(t, err)
			}
			if tt.cgroupMemTotalV2 != nil {
				err := afero.WriteFile(appFs, cgroupMemTotalPathV2, tt.cgroupMemTotalV2, 0644)
				assert.NoError(t, err)
			}
			memTotal = tt.total
			err := checkMemTotal()
			assert.NoError(t, err)
			err = checkCgroup()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, memTotal)
		})
	}
}
