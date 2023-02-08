package priority_queue

import "container/heap"

type Item[T any] struct {
	Value    T
	Priority int64
	Index    int
}
type PriorityQueue[T any] []*Item[T]

func NewPriorityQueue[T any](capacity int64) PriorityQueue[T] {
	return make(PriorityQueue[T], 0, capacity)
}
func (pq *PriorityQueue[T]) Less(i, j int) bool {
	return (*pq)[i].Priority < (*pq)[j].Priority
}
func (pq *PriorityQueue[T]) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
	(*pq)[i].Index, (*pq)[j].Index = i, j
}
func (pq *PriorityQueue[T]) Push(item interface{}) {
	l := len(*pq)
	c := cap(*pq)
	if l+1 > c {
		npq := make(PriorityQueue[T], l, c*2)
		copy(npq, *pq)
		*pq = npq
	}
	*pq = (*pq)[0 : l+1]
	tmp := (item).(*Item[T])
	(*pq)[l] = tmp
}
func (pq *PriorityQueue[T]) Pop() any {
	l := len(*pq)
	c := cap(*pq)
	if l < (c/2) && c > 32 {
		npq := make(PriorityQueue[T], l, c/2)
		copy(npq, *pq)
		*pq = npq
	}
	item := (*pq)[l-1]
	*pq = (*pq)[0 : l-1]
	return item
}
func (pq *PriorityQueue[T]) Len() int {
	return len(*pq)
}
func (pq *PriorityQueue[T]) PeekAndShift(max int64) (*Item[T], int64) {
	if pq.Len() == 0 {
		return nil, 0
	}
	item := (*pq)[0]
	if item.Priority > max {
		return nil, item.Priority - max
	}
	heap.Remove(pq, 0)
	return item, 0
}
