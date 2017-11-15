package mpdclient

func NewRequestQueue(size int) *RequestQueue {
	return &RequestQueue{
		entries: make([]*RequestQueueItem, size),
		size:    size,
		Count:   -1,
	}
}

func (q *RequestQueue) Push(entry *RequestQueueItem) bool {
	if q.Count == (q.size - 1) {
		return false
	}

	q.Count++
	q.entries[q.Count] = entry

	return true
}

func (q *RequestQueue) Pop() (*RequestQueueItem, bool) {
	if q.Count < 0 {
		return nil, false
	}
	value := q.entries[q.Count]
	q.entries[q.Count] = nil
	q.Count--

	return value, true
}

func (q *RequestQueue) Dump() map[int]string {
	response := map[int]string{}

	for i := 0; i < q.Count+1; i++ {
		response[i] = q.entries[i].Title
	}

	return response
}

func (q *RequestQueue) Has(title string) bool {
	for i := 0; i < q.Count+1; i++ {
		if q.entries[i].Title == title {
			return true
		}
	}
	return false
}
