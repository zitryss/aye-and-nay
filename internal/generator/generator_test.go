package generator

import (
	"flag"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	unit        = flag.Bool("unit", false, "")
	integration = flag.Bool("int", false, "")
	ci          = flag.Bool("ci", false, "")
)

func TestGenId(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	id, ids := GenId()
	id, ids = GenId()
	wg := sync.WaitGroup{}
	wg.Add(2)
	ch1 := make(chan struct{}, 1)
	ch2 := make(chan struct{}, 1)
	assert.NotPanics(t, func() {
		go func() {
			defer wg.Done()
			for i := 0; i < IDS_SPAN/2; i++ {
				_ = id()
			}
			ch1 <- struct{}{}
			<-ch2
			for i := 0; i < IDS_SPAN/2; i++ {
				_ = ids.Uint64(i)
			}
		}()
		go func() {
			defer wg.Done()
			for i := IDS_SPAN / 2; i < IDS_SPAN; i++ {
				_ = id()
			}
			ch2 <- struct{}{}
			<-ch1
			for i := IDS_SPAN / 2; i < IDS_SPAN; i++ {
				_ = ids.Uint64(i)
			}
		}()
	})
	wg.Wait()
	assert.Len(t, ids.logBook, IDS_SPAN)
	assert.Equal(t, ids.logBook[0], uint64(100))
	assert.Equal(t, ids.logBook[len(ids.logBook)-1], uint64(199))
	assert.Panics(t, func() { id() })
	assert.Panics(t, func() {
		_, ids := GenId()
		_ = ids.Base64(0)
	})
}
