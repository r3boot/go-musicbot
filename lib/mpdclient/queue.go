package mpdclient

import (
	"sync"
)

func NewRequestQueue(size int) *RequestQueue {
	return &RequestQueue{
		entries: make([]*RequestQueueItem, size),
		size:    size,
		count:   0,
		mutex:   sync.RWMutex{},
	}
}

func (q *RequestQueue) Size() int {
	return q.count
}

func (q *RequestQueue) Push(entry *RequestQueueItem) bool {
	if q.count == (q.size - 1) {
		return false
	}

	if q.Has(entry.Title) {
		return false
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.entries[q.count] = entry
	q.count++

	return true
}

func (q *RequestQueue) Pop() (*RequestQueueItem, bool) {
	var value *RequestQueueItem

	q.mutex.Lock()
	defer q.mutex.Unlock()

	value = nil

	switch {
	case q.count < 0:
		{
			return nil, false
		}
	case q.count == 0:
		{
			return nil, false
		}
	case q.count == 1:
		{
			value = q.entries[0]
			q.entries[0] = nil
		}
	default:
		{
			value = q.entries[0]
			q.entries[0] = nil
			for i := 1; i < q.count; i++ {
				q.entries[i-1] = q.entries[i]
				q.entries[i] = nil
			}
		}
	}

	q.count--

	return value, true
}

func (q *RequestQueue) Dump() map[int]string {
	response := map[int]string{}

	q.mutex.RLock()
	defer q.mutex.RUnlock()

	for i := 0; i < q.count; i++ {
		response[i] = q.entries[i].Title
	}

	return response
}

func (q *RequestQueue) Has(title string) bool {

	q.mutex.RLock()
	defer q.mutex.RUnlock()

	for i := 0; i < q.count; i++ {
		if q.entries[i].Title == title {
			return true
		}
	}
	return false
}
