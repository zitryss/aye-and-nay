package compressor

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zitryss/aye-and-nay/domain/model"
)

func TestImaginaryPositive(t *testing.T) {
	if !*integration {
		t.Skip()
	}
	tests := []struct {
		filename string
	}{
		{
			filename: "alan.jpg",
		},
		{
			filename: "dennis.png",
		},
		{
			filename: "big.jpg",
		},
		{
			filename: "tim.gif",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			im, err := NewImaginary(context.Background(), DefaultImaginaryConfig)
			require.NoError(t, err)
			b, err := os.ReadFile("../../testdata/" + tt.filename)
			assert.NoError(t, err)
			buf := bytes.NewBuffer(b)
			f := model.File{Reader: buf, Size: int64(buf.Len())}
			_, err = im.Compress(context.Background(), f)
			assert.NoError(t, err)
		})
	}
}

func TestImaginaryNegative(t *testing.T) {
	if !*integration {
		t.Skip()
	}
	if testing.Short() {
		t.Skip("short flag is set")
	}
	im, err := NewImaginary(context.Background(), DefaultImaginaryConfig)
	require.NoError(t, err)
	b, err := os.ReadFile("../../testdata/john.bmp")
	assert.NoError(t, err)
	buf := bytes.NewBuffer(b)
	f1 := model.File{Reader: buf, Size: int64(buf.Len())}
	f2, err := im.Compress(context.Background(), f1)
	assert.NoError(t, err)
	assert.Equal(t, f1.Size, f2.Size)
}

func TestImaginaryHealth(t *testing.T) {
	if !*integration {
		t.Skip()
	}
	im, err := NewImaginary(context.Background(), DefaultImaginaryConfig)
	require.NoError(t, err)
	_, err = im.Health(context.Background())
	assert.NoError(t, err)
}
