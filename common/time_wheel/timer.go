package time_wheel

import (
	"container/list"
	"sync/atomic"
	"unsafe"
)

// Timer as event, when timer time exceeded, task will be executed.
type Timer struct {
	expiration int64 // milliseconds
	task       func()
	b          unsafe.Pointer

	element *list.Element //
}

// get the bucket where timer belong
func (t *Timer) getBucket() *bucket {
	return (*bucket)(atomic.LoadPointer(&t.b))
}
func (t *Timer) setBucket(b *bucket) {
	atomic.StorePointer(&t.b, unsafe.Pointer(b))
}

// Stop the timer
func (t *Timer) Stop() bool {
	stopped := false
	for b := t.getBucket(); b != nil; b = t.getBucket() {
		stopped = b.Remove(t)
	}
	return stopped
}
