package mpdclient

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gompd/mpd"
)

func (m *MPDClient) Connect() error {
	var err error

	if m.config.MPD.Password != "" {
		m.conn, err = mpd.DialAuthenticated("tcp", m.address, m.config.MPD.Password)
		if err != nil {
			return fmt.Errorf("MPDClient.Connect mpd.DialAuthenticated: %v", err)
		}
	} else {
		m.conn, err = mpd.Dial("tcp", m.address)
		if err != nil {
			return fmt.Errorf("MPDClient.Connect mpd.Dial: %v", err)
		}
	}

	return nil
}

func (m *MPDClient) KeepAlive() {
	var err error

	for {
		if m.conn == nil { // Socket is closed, connect to mpd again
			if err = m.Connect(); err != nil {
				time.Sleep(time.Second * 10)
				continue
			}
		}

		if err = m.conn.Ping(); err != nil { // Ping command failed, reconnect to mpd
			m.Close()
			if err = m.Connect(); err != nil {
				time.Sleep(time.Second * 10)
				continue
			}
		}

		time.Sleep(time.Second * 3)
	}
}

func (m *MPDClient) MaintainMPDState() {
	for {
		curSongData := NowPlayingData{}

		songAttrs, err := m.conn.CurrentSong()
		if err != nil {
			log.Warningf("MPDClient.MaintainMPDState m.conn.CurrentSong: %v", err)
			return
		}

		fileName := songAttrs["file"]
		fullPath := m.mp3.BaseDir + "/" + fileName
		curSongData.Title = fileName[:len(fileName)-16]

		statusAttrs, err := m.conn.Status()
		if err != nil {
			log.Warningf("MPDClient.MaintainMPDState m.conn.Status: %v", err)
			return
		}
		curSongData.Elapsed, err = strconv.ParseFloat(statusAttrs["elapsed"], 32)
		if err != nil {
			log.Warningf("MPDClient.MaintainMPDState strconv.ParseFloat: %s: %v", statusAttrs["elapsed"], err)
			return
		}

		curSongData.Duration, err = strconv.ParseFloat(statusAttrs["duration"], 32)
		if err != nil {
			log.Warningf("MPDClient.MaintainMPDState strconv.ParseFloat: %s: %v", statusAttrs["elapsed"], err)
			return
		}

		curSongData.Remaining = curSongData.Duration - curSongData.Elapsed

		curSongData.Rating = m.mp3.GetRating(fullPath)

		m.np = curSongData

		m.UpdateQueueIDs()

		time.Sleep(1 * time.Second)
	}

}

func findID(entries []mpd.Attrs, title string) int {
	for _, entry := range entries {
		if !strings.HasPrefix(entry["file"], title) {
			continue
		}

		id, err := strconv.Atoi(entry["Id"])
		if err != nil {
			log.Warningf("findID: Failed to convert int: %v", err)
			break
		}

		return id
	}
	return -1
}

func (m *MPDClient) UpdateQueueIDs() {
	var i int

	queueEntries := m.queue.Dump()

	if m.queue.count > 0 && m.np.Title == queueEntries[0] {
		m.queue.Pop()
		queueEntries = m.queue.Dump()
	}

	playlist, err := m.conn.PlaylistInfo(-1, -1)
	if err != nil {
		log.Warningf("MPDClient.UpdateQueueIDs m.conn.PlaylistInfo: %v", err)
		return
	}

	maxCount := 9
	if m.queue.count < maxCount {
		maxCount = m.queue.count
	}

	for i = 0; i < maxCount; i++ {
		title := queueEntries[i]
		id := findID(playlist, title)
		if id == -1 {
			log.Warningf("id for %s not found", title)
			continue
		}
		prio := 9 - i
		err := m.conn.PrioId(prio, id)
		if err != nil {
			log.Warningf("MPDClient.UpdateQueueIDs m.conn.PrioId: %s: %v", title, err)
		}
	}
}

func (m *MPDClient) UpdateDB(fname string) error {
	_, err := m.conn.Update(fname)
	time.Sleep(1 * time.Second)
	m.Add(fname)
	return err
}

func (m *MPDClient) Close() error {
	var err error

	if err = m.conn.Close(); err != nil {
		return fmt.Errorf("MPDClient.Close m.conn.Close: %v", err)
	}

	return nil
}

func (m *MPDClient) NowPlaying() string {
	attrs, err := m.conn.CurrentSong()
	if err != nil {
		return fmt.Sprintf("MPDClient.NowPlaying m.conn.CurrentSong: %v", err)
	}
	return attrs["file"]
}

func (m *MPDClient) Duration() string {
	attrs, err := m.conn.CurrentSong()
	if err != nil {
		return fmt.Sprintf("MPDClient.Duration m.conn.CurrentSong: Failed to fetch current song info: %v", err)
	}

	rawDuration := strings.Split(attrs["duration"], ".")[0]
	rawDuration += "s"
	duration, err := time.ParseDuration(rawDuration)
	if err != nil {
		return fmt.Sprintf("MPDClient.Duration time.ParseDuration: Failed to parse duration: %v", err)
	}
	return duration.String()
}

func (m *MPDClient) Next() string {
	m.conn.Next()
	return m.NowPlaying()
}

func (m *MPDClient) Play() string {
	m.Shuffle()

	m.conn.Play(1)
	return m.NowPlaying()
}

func (m *MPDClient) PlayPos(pos int) string {
	m.conn.Play(pos)
	return m.NowPlaying()
}

func (m *MPDClient) Shuffle() {
	m.conn.Shuffle(-1, -1)
}

func (m *MPDClient) Add(fileName string) {
	log.Infof("Adding %s to playlist", fileName)
	m.conn.Add(fileName)
}

func (m *MPDClient) TypeAheadQuery(q string) []string {
	result, err := m.conn.Search("filename", q)
	if err != nil {
		errmsg := fmt.Sprintf("MPDClient.TypeAheadQuery m.conn.Search: %v", err)
		log.Warningf(errmsg)
		return nil
	}

	foundFiles := []string{}
	for _, entry := range result {
		foundFiles = append(foundFiles, entry["file"][:len(entry["file"])-16])
	}

	return foundFiles
}

func (m *MPDClient) Search(q string) (int, error) {
	result, err := m.conn.Search("filename", q)
	if err != nil {
		return -1, fmt.Errorf("MPDClient.Search m.conn.Search: %v", err)
	}

	if len(result) == 0 {
		return -1, fmt.Errorf("MPDClient.Search: no songs found")
	}

	fileName := result[0]["file"]

	curPlaylist, err := m.conn.PlaylistInfo(-1, -1)
	if err != nil {
		return -1, fmt.Errorf("MPDClient.Search m.conn.PlaylistInfo: %v", err)
	}

	for _, song := range curPlaylist {
		if song["file"] == fileName {
			pos, err := strconv.Atoi(song["Pos"])
			if err != nil {
				return -1, fmt.Errorf("MPDClient.Search strconv.Atoi: %v", err)
			}
			return pos, nil
		}
	}

	return -1, fmt.Errorf("MPDClient.Search: failed to search mpd")
}

func (m *MPDClient) Enqueue(query string) (int, error) {

	pos, err := m.Search(query)
	if err != nil {
		return -1, fmt.Errorf("MPDClient.Enqueue: %v", err)
	}

	title, err := m.GetTitle(pos)
	if err != nil {
		return -1, fmt.Errorf("MPDClient.Enqueue: %v", err)
	}
	title = title[:len(title)-16]

	if m.queue.Has(title) {
		return -1, fmt.Errorf("MPDClient.Enqueue: already enqueued")
	}

	if m.queue.count >= MAX_QUEUE_ITEMS {
		return -1, fmt.Errorf("MPDClient.Enqueue: queue is full")
	}

	qitem := &RequestQueueItem{
		Title: title,
		Pos:   pos,
	}

	ok := m.queue.Push(qitem)
	if !ok {
		return -1, fmt.Errorf("MPDClient.Enqueue m.queue.Push: Failed to push")
	}

	return pos, nil
}

func (m *MPDClient) GetPlayQueue() (map[int]string, error) {
	return m.queue.Dump(), nil
}

func (m *MPDClient) GetTitle(pos int) (string, error) {
	info, err := m.conn.PlaylistInfo(pos, -1)
	if err != nil {
		return "", fmt.Errorf("MPDClient.GetTitle m.conn.PlaylistInfo: %v", err)
	}

	return info[0]["file"], nil
}
