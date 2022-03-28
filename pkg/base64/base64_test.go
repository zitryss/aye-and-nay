package base64

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func FuzzFromUint64(f *testing.F) {
	f.Add(uint64(0))
	f.Fuzz(func(t *testing.T, u1 uint64) {
		b64 := FromUint64(u1)
		u2, err := ToUint64(b64)
		assert.NoError(t, err)
		assert.Equal(t, u1, u2)
	})
}

func FuzzToUint64(f *testing.F) {
	f.Add("AAAAAAAAAAA")
	f.Fuzz(func(t *testing.T, s string) {
		_, err := ToUint64(s)
		if err != nil {
			t.Skip()
		}
	})
}
