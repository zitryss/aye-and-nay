package generator

import (
	"fmt"
	"sync"

	"github.com/zitryss/aye-and-nay/pkg/base64"
)

const (
	IDS_SPAN = 100
)

var (
	m       sync.Mutex
	indexId uint64
)

func GenId() (func() uint64, *syncLogBook) {
	lb := syncLogBook{m: sync.Mutex{}, logBook: map[int]uint64{}}
	fn := func() func() uint64 {
		m.Lock()
		firstId := indexId
		indexId += IDS_SPAN
		m.Unlock()
		mFn := sync.Mutex{}
		i := -1
		curId := firstId - 1
		lastId := firstId + IDS_SPAN - 1
		return func() uint64 {
			mFn.Lock()
			i++
			curId++
			if curId > lastId {
				panic("id out of bounds")
			}
			lb.set(i, curId)
			id := curId
			mFn.Unlock()
			return id
		}
	}
	return fn(), &lb
}

type syncLogBook struct {
	m       sync.Mutex
	logBook map[int]uint64
}

func (lb *syncLogBook) set(i int, id uint64) {
	lb.m.Lock()
	defer lb.m.Unlock()
	lb.logBook[i] = id
}

func (lb *syncLogBook) Uint64(i int) uint64 {
	lb.m.Lock()
	defer lb.m.Unlock()
	id, ok := lb.logBook[i]
	if !ok {
		panic(fmt.Sprintf("id #%d not found", i))
	}
	return id
}

func (lb *syncLogBook) Base64(i int) string {
	return base64.FromUint64(lb.Uint64(i))
}
