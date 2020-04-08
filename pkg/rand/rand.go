package rand

import (
	crand "crypto/rand"
	"encoding/base64"
	"math"
	mrand "math/rand"
)

var (
	Id = func(length int) (string, error) {
		n := int(math.Ceil(float64(length) * 3 / 4))
		b := make([]byte, n)
		_, err := crand.Read(b)
		if err != nil {
			return "", err
		}
		str := base64.RawURLEncoding.EncodeToString(b)
		if len(str) > length {
			str = str[:len(str)-1]
		}
		return str, nil
	}

	Shuffle = mrand.Shuffle
)
