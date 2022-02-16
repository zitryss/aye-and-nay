package base64

import (
	"encoding/binary"

	"github.com/segmentio/asm/base64"
)

func FromUint64(u uint64) string {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, u)
	return base64.RawURLEncoding.EncodeToString(b)
}

func ToUint64(s string) (uint64, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return 0x0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}
