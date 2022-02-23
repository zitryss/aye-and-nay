package generator

import (
	"fmt"
	"sync"

	"github.com/zitryss/aye-and-nay/pkg/base64"
)

const (
	span = 100
)

var (
	m       sync.Mutex
	indexId uint64
)

type IdGenFunc func() uint64

type Ids interface {
	Uint64(i int) uint64
	Base64(i int) string
}

func GenId() (IdGenFunc, *IdLogBook) {
	lb := IdLogBook{m: sync.Mutex{}, logBook: map[int]uint64{}, valid: true}
	fn := func() IdGenFunc {
		m.Lock()
		firstId := indexId
		indexId += span
		m.Unlock()
		mFn := sync.Mutex{}
		i := -1
		curId := firstId - 1
		lastId := firstId + span - 1
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

type IdLogBook struct {
	m       sync.Mutex
	logBook map[int]uint64
	valid   bool
}

func (lb *IdLogBook) set(i int, id uint64) {
	if lb == nil || !lb.valid {
		return
	}
	lb.m.Lock()
	defer lb.m.Unlock()
	_, ok := lb.logBook[i]
	if ok {
		lb.logBook[i] = id
	}
}

func (lb *IdLogBook) Uint64(i int) uint64 {
	if lb == nil || !lb.valid {
		return 0x0
	}
	lb.m.Lock()
	defer lb.m.Unlock()
	id, ok := lb.logBook[i]
	if !ok {
		panic(fmt.Sprintf("id #%d not found", i))
	}
	return id
}

func (lb *IdLogBook) Base64(i int) string {
	return base64.FromUint64(lb.Uint64(i))
}
