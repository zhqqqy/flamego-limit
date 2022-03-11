package limit

import (
	"context"
	"github.com/flamego/cache"
	"github.com/flamego/flamego"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type limit struct {
	pool       sync.Pool
	storage    cache.Cache
	max        int
	expiration uint64
}
type window struct {
	prevHits int
	currHits int
	endTime  uint64
}
type handle = func(c flamego.Context) bool

func newLimit(opt Options) Limit {
	initer := cache.MemoryIniter()
	mem, err := initer(context.Background())
	if err != nil {
		panic("cache: " + err.Error())
	}
	return &limit{
		pool: sync.Pool{
			New: func() interface{} {
				return new(window)
			},
		},
		storage:    mem,
		max:        opt.Max,
		expiration: uint64(opt.Expiration.Seconds()),
	}
}

func (l *limit) DoLimit(key string) handle {
	var (
		timestamp = uint64(time.Now().Unix())
		mux       = &sync.RWMutex{}
	)

	// Update timestamp every second
	go func() {
		for {
			atomic.StoreUint64(&timestamp, uint64(time.Now().Unix()))
			time.Sleep(1 * time.Second)

		}
	}()

	return func(c flamego.Context) bool {
		mux.Lock()
		it, _ := l.storage.Get(context.Background(), key)
		if it == nil {
			it = l.pool.Get()
		}
		wd := it.(*window)
		ts := atomic.LoadUint64(&timestamp)
		if wd.endTime == 0 { //
			wd.endTime = ts + l.expiration
		} else if ts >= wd.endTime { //Time has passed in front of the window
			wd.prevHits = wd.currHits //Set the number of hit the front window
			wd.currHits = 0           //The number of hits to reset the current window
			elapsed := ts - wd.endTime
			if elapsed > l.expiration { // If it has exceeded the time of two windows, set the time of the next window to start from this hit data + the window is often, otherwise it is subtracted
				wd.endTime = ts + l.expiration
			} else {
				wd.endTime = ts + l.expiration - elapsed
			}
		}
		wd.currHits++
		// Calculate when it resets in seconds
		resetInSec := wd.endTime - ts

		weight := float64(resetInSec) / float64(l.expiration)
		rate := int(float64(wd.prevHits)*weight) + wd.currHits
		remaining := l.max - rate
		log.Println(remaining)
		//Because we need to get the number of hits of the previous window, we set the expiration time plus remaining as the window's expiration time
		l.storage.Set(context.Background(), key, wd, time.Duration(resetInSec+l.expiration)*time.Second)
		mux.Unlock()
		if remaining < 0 {
			return true
		}
		return false
	}

}
