package delay_queue

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"tools/common/priority_queue"
)

type DelayQueue[T any] struct {
	pq         priority_queue.PriorityQueue[T]
	mu         sync.Mutex
	sleeping   int32
	wakeupChan chan struct{}
	OutChan    chan T
}

func NewDelayQueue[T any](gctx context.Context, nowFunc func() int64, size int64) *DelayQueue[T] {
	dq := &DelayQueue[T]{
		pq:         priority_queue.NewPriorityQueue[T](size),
		wakeupChan: make(chan struct{}),
		OutChan:    make(chan T, size),
	}
	go func() { dq.Poll(gctx, nowFunc) }()
	return dq
}
func (dq *DelayQueue[T]) Offer(value T, expiration int64) {
	item := &priority_queue.Item[T]{
		Value:    value,
		Priority: expiration,
	}
	dq.mu.Lock()
	heap.Push(&dq.pq, item)
	idx := item.Index
	dq.mu.Unlock()

	if idx == 0 {
		if atomic.CompareAndSwapInt32(&dq.sleeping, 1, 0) {
			dq.wakeupChan <- struct{}{}
		}
	}
}

func (dq *DelayQueue[T]) Poll(gctx context.Context, nowFunc func() int64) {
	for {
		now := nowFunc()
		dq.mu.Lock()

		item, delta := dq.pq.PeekAndShift(now)

		if item == nil {
			//have no timer in the heap , sleeping the dq
			atomic.StoreInt32(&dq.sleeping, 1)
		}
		dq.mu.Unlock()
		if item == nil {
			if delta == 0 {
				//no timer in the dq, waiting for wakeup or exit
				select {
				case <-dq.wakeupChan:
					continue
				case <-gctx.Done():
					return
				}
			} else if delta > 0 {
				fmt.Println("delta = ", delta)
				select {
				case <-dq.wakeupChan:
					fmt.Println("A near timer coming.")
					// already sleeping, add a new nearest timer, wakeup the dq
					continue
				case <-time.After(time.Duration(delta) * time.Millisecond):
					//set a timer for waiting this delta
					if atomic.SwapInt32(&dq.sleeping, 0) == 0 {
						//already awake, but not consume the wakeupChan yet
						<-dq.wakeupChan
					}
					continue
				case <-gctx.Done():
					return
				}
			}
		}
		select {
		case dq.OutChan <- item.Value:
			//push into out chan
		case <-gctx.Done():
			//exit
			return
		}
	}
}
