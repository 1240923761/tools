package time_wheel

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
	"tools/common/delay_queue"
	"unsafe"
)

// TimeWheel is a Leveled Time wheel
type TimeWheel[T any] struct {
	tick      int64 //bucket time level,milliseconds
	wheelSize int64 //number of buckets

	interval    int64 //this wheel's time level, tick*wheelSize
	currentTime int64
	buckets     []*bucket //list of buckets
	queue       *delay_queue.DelayQueue[*bucket]

	overflowWheel unsafe.Pointer //

	gctx      context.Context
	waitGroup sync.WaitGroup
}

func NewTimeWheel[T any](gctx context.Context, tickLevel time.Duration, wheelSize int64) *TimeWheel[T] {

	tickMs := int64(tickLevel / time.Millisecond)
	//fmt.Println(tickMs)
	if tickMs <= 0 {
		panic(errors.New("tick must be greater than or equal to 1ms"))
	}

	//
	startMs := time.Now().UnixMilli()

	return newTimeWheel[T](
		gctx,
		tickMs,
		wheelSize,
		startMs,
		delay_queue.NewDelayQueue[*bucket](gctx, NowFunc, wheelSize),
	)

}
func newTimeWheel[T any](gctx context.Context, tickMs int64, wheelSize int64, startMs int64, queue *delay_queue.DelayQueue[*bucket]) *TimeWheel[T] {
	buckets := make([]*bucket, wheelSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	return &TimeWheel[T]{
		tick:        tickMs,
		wheelSize:   wheelSize,
		currentTime: Truncate(startMs, tickMs),
		interval:    tickMs * wheelSize,
		buckets:     buckets,
		queue:       queue,
		gctx:        gctx,
	}
}
func NowFunc() int64 {
	return time.Now().UnixMilli()
}
func Truncate(a, b int64) int64 {
	timeNow := time.UnixMilli(a)
	level := time.Duration(b) * time.Millisecond
	return timeNow.Truncate(level).UnixMilli()
}

//add timer to bucket

func (tw *TimeWheel[T]) add(t *Timer) bool {
	currentTime := atomic.LoadInt64(&tw.currentTime)
	if t.expiration < currentTime+tw.tick {
		//timer expiration already exceeded
		return false
	} else if t.expiration < currentTime+tw.interval {
		virturalID := t.expiration / tw.tick
		b := tw.buckets[virturalID%tw.wheelSize]
		b.Add(t)

		if b.SetExpiration(virturalID * tw.tick) {
			tw.queue.Offer(b, b.Expiration())
		}
		return true
	} else {
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel == nil {
			atomic.CompareAndSwapPointer(
				&tw.overflowWheel,
				nil,
				unsafe.Pointer(
					newTimeWheel[T](
						tw.gctx,
						tw.interval,
						tw.wheelSize,
						currentTime,
						tw.queue,
					),
				),
			)
		}
		time.Sleep(1 * time.Second)
		overflowWheel = atomic.LoadPointer(&tw.overflowWheel)
		return (*TimeWheel[T])(overflowWheel).add(t)
	}
}

func (tw *TimeWheel[T]) addOrRun(t *Timer) {
	if !tw.add(t) {
		go t.task()
	}
}

func (tw *TimeWheel[T]) advanceClock(expiration int64) {
	currentTime := atomic.LoadInt64(&tw.currentTime)

	if expiration >= currentTime+tw.tick {
		currentTime = Truncate(expiration, tw.tick)
		atomic.StoreInt64(&tw.currentTime, currentTime)

		//overflow timeWheel change current
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel != nil {
			(*TimeWheel[T])(overflowWheel).advanceClock(currentTime)
		}
	}
}

func (tw *TimeWheel[T]) Start() {
	go func() {
		for {
			select {
			case elem := <-tw.queue.OutChan:
				tw.advanceClock(elem.Expiration())
				elem.Flush(tw.addOrRun)
			case <-tw.gctx.Done():
				return
			}
		}
	}()
}
func (tw *TimeWheel[T]) Stop() {

}
func (tw *TimeWheel[T]) AfterFunc(d time.Duration, f func()) *Timer {
	t := &Timer{
		expiration: time.Now().Add(d).UnixMilli(),
		task:       f,
	}
	tw.addOrRun(t)
	return t
}
