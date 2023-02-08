package time_wheel

import (
	"container/list"
	"sync"
	"sync/atomic"
)

type bucket struct {
	expiration int64 //milliseconds
	mu         sync.Mutex
	timers     *list.List
}

func newBucket() *bucket {
	return &bucket{
		expiration: -1,
		timers:     list.New(),
	}
}
func (b *bucket) Expiration() int64 {
	return atomic.LoadInt64(&b.expiration)
}
func (b *bucket) SetExpiration(expiration int64) bool {
	return atomic.SwapInt64(&b.expiration, expiration) != expiration
}
func (b *bucket) Add(t *Timer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	e := b.timers.PushBack(t)
	t.setBucket(b)
	t.element = e
}

// delete timer
func (b *bucket) remove(t *Timer) bool {
	if t.getBucket() != b {
		// if this timer not belong this bucket, return false
		return false
	}
	b.timers.Remove(t.element)
	t.setBucket(nil)
	t.element = nil
	return true
}
func (b *bucket) Remove(t *Timer) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.remove(t)
}
func (b *bucket) Flush(reinsert func(*Timer)) {
	b.mu.Lock()
	var ts = make([]*Timer, 0, b.timers.Len())
	for e := b.timers.Front(); e != nil; {
		next := e.Next()
		t := e.Value.(*Timer)
		b.remove(t)
		ts = append(ts, t)
		e = next
	}
	b.mu.Unlock()
	b.SetExpiration(-1)

	for _, t := range ts {
		reinsert(t)
	}
}
