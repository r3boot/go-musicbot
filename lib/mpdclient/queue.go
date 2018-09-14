package mpdclient

import (
	"fmt"
	"sort"
)

const (
	MAX_QUEUE_ITEMS = 8
)

var (
	queueSize = 0
)

func (q PlayQueueEntries) ToMap() PlayQueueEntries {
	entries := PlayQueueEntries{}

	for _, entry := range q {
		entries[entry.QPrio] = entry
	}

	return entries
}

func (q PlayQueueEntries) Keys() []int {
	allEntries := []int{}
	for idx, _ := range q {
		allEntries = append(allEntries, idx)
	}

	sort.Ints(allEntries)
	return allEntries
}

func (q PlayQueueEntries) Has(title string) bool {
	for _, entry := range q {
		if entry.Title == title {
			return true
		}
	}

	return false
}

func (q PlayQueue) GetAll() PlayQueueEntries {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return q.entries
}

func (q PlayQueue) Size() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.entries)
}

func (q PlayQueue) HasId(id int) bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	for _, entry := range q.entries {
		if entry.Id == id {
			return true
		}
	}
	return false
}

func (q PlayQueue) Add(id int, entry *PlaylistEntry) error {
	if q.Size() >= MAX_QUEUE_ITEMS {
		return fmt.Errorf("PlayQueue.Add: queue is full")
	}

	if q.HasId(id) {
		return fmt.Errorf("PlayQueue.Add: id already queued")
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.entries[id] = entry

	return nil
}

func (q PlayQueue) Delete(id int) error {
	if !q.HasId(id) {
		return fmt.Errorf("PlayQueue.Delete: no such id")
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.conn.PrioId(0, id)
	delete(q.entries, id)

	for id, _ := range q.entries {
		if q.entries[id].QPrio > 0 {
			q.entries[id].QPrio -= 1
		}
		mpdPrio := 9 - q.entries[id].QPrio
		q.conn.PrioId(mpdPrio, id)
	}

	return nil
}

func (m *MPDClient) FirstFreeQueueSlot(startIdx int) int {
	for i := startIdx; i < MAX_QUEUE_ITEMS; i++ {
		_, ok := m.np.RequestQueue[i]
		if !ok {
			return i
		}
	}

	return -1
}

func (m *MPDClient) UpdateQueueIDs() error {
	for _, entry := range m.queue.GetAll() {
		if entry.Id == m.np.Id {
			err := m.queue.Delete(entry.Id)
			if err != nil {
				return fmt.Errorf("MPDClient.UpdateQueueIDs: %v", err)
			}
			log.Debugf("MPDClient.UpdateQueueIDs: removed %d from queue", entry.Id)
			continue
		}
	}

	log.Debugf("Request queue has %d items", m.queue.Size())
	return nil
}

func (m *MPDClient) Enqueue(query string) (*PlaylistEntry, error) {
	if len(m.queue.entries) >= MAX_QUEUE_ITEMS-1 {
		return nil, fmt.Errorf("MPDClient.Enqueue: queue is full")
	}



	/*
	pos, err := m.Search(query)
	if err != nil {
		return nil, fmt.Errorf("MPDClient.Enqueue: %v", err)
	}

	fname, err := m.GetTitle(pos)
	if err != nil {
		return nil, fmt.Errorf("MPDClient.Enqueue: %v", err)
	}
	*/

	playList, err := m.GetPlaylist()
	if err != nil {
		return nil, fmt.Errorf("MPDClient.Enqueue: %v", err)
	}

	entry, err := playList.GetEntryByPos(pos)
	if err != nil {
		return nil, fmt.Errorf("MPDClient.Enqueue: %v", err)
	}

	allTags, err := m.id3.GetTags()
	if err != nil {
		return nil, fmt.Errorf("MPDClient.Enqueue: %v", err)
	}

	for tagFname, meta := range allTags {
		if tagFname != fname {
			continue
		}
		entry.Artist = meta.Artist
		entry.Title = meta.Title
	}

	entry.QPrio = m.queue.Size()

	err = m.PrioId(entry.Id, entry.QPrio)
	if err != nil {
		err = fmt.Errorf("MPDClient.UpdateQueueIDs: failed to set prio: %v", err)
		log.Warningf("%v", err)
		return nil, err
	}
	log.Debugf("MPDClient.UpdateQueueIDs: prio for trackId %d set to %d", entry.Id, 9-entry.QPrio)

	err = m.queue.Add(entry.Id, entry)
	if err != nil {
		err = fmt.Errorf("MPDClient.UpdateQueueIDs: %v", err)
		log.Warningf("%v", err)
		return nil, err
	}

	return entry, nil
}

func (m *MPDClient) GetPlayQueue() PlayQueueEntries {
	allEntries := m.queue.GetAll().ToMap()
	return allEntries
}

func (m *MPDClient) PreLoadPlayQueue() error {
	playlist, err := m.GetPlaylist()
	if err != nil {
		return fmt.Errorf("MPDClient.PreLoadPlayQueue: %v", err)
	}

	preloadQueue := PlayQueueEntries{}

	for _, entry := range playlist {
		if entry.prio > 0 {
			entry.QPrio = 9 - entry.prio
			preloadQueue[entry.Id] = entry
		}
	}

	m.queue.entries = preloadQueue
	m.np.RequestQueue = preloadQueue

	log.Debugf("MPDClient.PreLoadPlayQueue: preloaded %d items from mpd playlist", len(preloadQueue))

	return nil
}
