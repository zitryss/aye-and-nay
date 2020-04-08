package retry

import (
	"time"
)

func Do(times int, pause time.Duration, fn func() error) error {
	n := -1
	for {
		err := fn()
		if err == nil {
			return nil
		}
		n++
		if n >= times {
			return err
		}
		time.Sleep(pause)
	}
}
