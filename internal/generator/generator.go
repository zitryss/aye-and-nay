package generator

import (
	"sync"
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

func (lb *syncLogBook) Get(i int) uint64 {
	lb.m.Lock()
	defer lb.m.Unlock()
	return lb.logBook[i]
}

func (lb *syncLogBook) set(i int, id uint64) {
	lb.m.Lock()
	defer lb.m.Unlock()
	lb.logBook[i] = id
}
