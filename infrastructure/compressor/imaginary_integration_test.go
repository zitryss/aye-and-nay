//go:build integration

package compressor

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/zitryss/aye-and-nay/domain/model"
)

func TestImaginaryPositive(t *testing.T) {
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
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile("../../testdata/" + tt.filename)
			if err != nil {
				t.Error(err)
			}
			buf := bytes.NewBuffer(b)
			f := model.File{
				Reader: buf,
				Size:   int64(buf.Len()),
			}
			_, err = im.Compress(context.Background(), f)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestImaginaryNegative(t *testing.T) {
	if testing.Short() {
		t.Skip("short flag is set")
	}
	im, err := NewImaginary(context.Background(), DefaultImaginaryConfig)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile("../../testdata/john.bmp")
	if err != nil {
		t.Error(err)
	}
	buf := bytes.NewBuffer(b)
	f1 := model.File{
		Reader: buf,
		Size:   int64(buf.Len()),
	}
	f2, err := im.Compress(context.Background(), f1)
	if err != nil {
		t.Error(err)
	}
	if f1.Size != f2.Size {
		t.Error("f1.Size != f2.Size")
	}
}

func TestImaginaryHealth(t *testing.T) {
	im, err := NewImaginary(context.Background(), DefaultImaginaryConfig)
	if err != nil {
		t.Fatal(err)
	}
	_, err = im.Health(context.Background())
	if err != nil {
		t.Error(err)
	}
}
