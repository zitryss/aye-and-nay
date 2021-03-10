package rand

import (
	"crypto/rand"
	"encoding/binary"
)

func Id() (uint64, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return 0x0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}
