package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/r3boot/test/lib/liquidsoap"
	"github.com/sirupsen/logrus"

	"github.com/r3boot/test/lib/dbclient"
)

const (
	MaxQueueSize = 10
)

type PlayQueue struct {
	items map[int]*dbclient.Track
	mux   sync.RWMutex
	ls    *liquidsoap.LSClient
}

func NewPlayQueue(client *liquidsoap.LSClient) *PlayQueue {
	queue := &PlayQueue{
		items: make(map[int]*dbclient.Track, MaxQueueSize),
		mux:   sync.RWMutex{},
		ls:    client,
	}

	go queue.updateQueueEntries()

	return queue
}

func (q *PlayQueue) Has(id int) bool {
	for _, track := range q.items {
		if track.Id == id {
			return true
		}
	}

	return false
}

func (q *PlayQueue) Size() int {
	return len(q.items)
}

func (q *PlayQueue) Add(track *dbclient.Track) error {
	if q.Size() >= MaxQueueSize {
		return fmt.Errorf("Queue is full")
	}

	/*
		if q.Has(track.Id) {
			return fmt.Errorf("Already queued")
		}
	*/

	newPrio := q.Size()

	q.mux.Lock()
	defer q.mux.Unlock()

	q.items[newPrio] = track
	/*
		err := q.mpd.SetPriority(track.Id, MaxQueueSize-newPrio)
		if err != nil {
			return fmt.Errorf("mpd.SetPriority: %v", err)
		}
	*/

	return nil
}

func (q *PlayQueue) Delete(id int) error {
	if !q.Has(id) {
		return fmt.Errorf("No such id")
	}

	if !q.Has(id) {
		return fmt.Errorf("No such id")
	}

	q.mux.Lock()
	defer q.mux.Unlock()

	// q.mpd.SetPriority(id, 0)
	delete(q.items, id)

	for _, entry := range q.items {
		if entry.Priority > 0 {
			entry.Priority -= 1
		}
		// q.mpd.SetPriority(entry.Id, MaxQueueSize-entry.Priority)
	}

	return nil
}

func (q *PlayQueue) updateQueueEntries() {
	log := logrus.WithFields(logrus.Fields{
		"module":   "PlayQueue",
		"function": "updateQueueEntries",
	})

	for {
		_, err := q.ls.GetQueue()
		if err != nil {
			log.Warningf("Failed to fetch queue info: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		time.Sleep(1000 * time.Millisecond)
	}

}
