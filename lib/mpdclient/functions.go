package mpdclient

import (
	"fmt"
	"time"

	"encoding/json"
	"strconv"

	"github.com/r3boot/go-musicbot/lib/albumart"
	"gompd/mpd"
	"path"
	"sort"
)

func (p *Playlist) ToJSON() ([]byte, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("Playlist.ToJSON json.Marshal: %v", err)
	}

	return data, nil
}

func (p *Playlist) ToArray() []*PlaylistEntry {
	tmpList := []*PlaylistEntry{}

	for _, entry := range *p {
		tmpList = append(tmpList, entry)
	}

	return tmpList
}

func (p *Playlist) HasFilename(fname string) bool {
	for _, entry := range *p {
		if entry.Filename == fname {
			return true
		}
	}

	return false
}

func (p *Playlist) GetEntryByPos(pos int) (*PlaylistEntry, error) {
	for _, entry := range *p {
		if entry.Pos == pos {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("Playlist.GetEntryByPos: entry not found in playlist")
}

func (a Artists) ToJSON() ([]byte, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("Artists.ToJSON json.Marshal: %v", err)
	}

	return data, nil
}

func (a Artists) Has(name string) bool {
	for _, entry := range a {
		if entry == name {
			return true
		}
	}

	return false
}

func (a Artists) Len() int           { return len(a) }
func (a Artists) Less(i, j int) bool { return a[i] < a[j] }
func (a Artists) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (m *MPDClient) Connect() error {
	var err error

	if m.password != "" {
		m.conn, err = mpd.DialAuthenticated("tcp", m.address, m.password)
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

func (m *MPDClient) Close() error {
	var err error

	if err = m.conn.Close(); err != nil {
		return fmt.Errorf("MPDClient.Close m.conn.Close: %v", err)
	}

	return nil
}

func (m *MPDClient) MaintainMPDState() {
	for {
		time.Sleep(1 * time.Second)

		tStart := time.Now()

		curSongData := NowPlayingData{}

		songAttrs, err := m.conn.CurrentSong()
		if err != nil {
			log.Warningf("MPDClient.MaintainMPDState m.conn.CurrentSong: %v", err)
			continue
		}

		fileName := songAttrs["file"]
		curSongData.Filename = fileName
		curSongData.Title = fileName[:len(fileName)-16]

		if m.curFile != fileName {
			m.curFile = fileName
			m.imageUrl = ""
		}

		curSongData.Id, err = strconv.Atoi(songAttrs["Id"])
		if err != nil {
			log.Warningf("MPDClient.MaintainMPDState strconv.Atoi: %v", err)
			continue
		}

		statusAttrs, err := m.conn.Status()
		if err != nil {
			log.Warningf("MPDClient.MaintainMPDState m.conn.Status: %v", err)
			continue
		}
		curSongData.Elapsed, err = strconv.ParseFloat(statusAttrs["elapsed"], 32)
		if err != nil {
			log.Warningf("MPDClient.MaintainMPDState strconv.ParseFloat: %s: %v", statusAttrs["elapsed"], err)
			continue
		}

		curSongData.Duration, err = strconv.ParseFloat(statusAttrs["duration"], 32)
		if err != nil {
			log.Warningf("MPDClient.MaintainMPDState strconv.ParseFloat: %s: %v", statusAttrs["elapsed"], err)
			continue
		}

		curSongData.Remaining = curSongData.Duration - curSongData.Elapsed

		curSongData.Rating, err = m.id3.GetRating(fileName)
		if err != nil {
			log.Warningf("MPDClient.MaintainMPDState: %v", err)
			continue
		}

		if m.imageUrl == "" {
			imgUrl, err := m.art.GetAlbumArt(curSongData.Title)
			if err != nil {
				log.Warningf("MPDClient.MaintainMPDState: %v", err)
			}

			if imgUrl != "" {
				m.imageUrl = imgUrl
			} else {
				m.imageUrl = albumart.NOTFOUND_URI
			}
		}

		err = m.UpdateQueueIDs()
		if err != nil {
			log.Warningf("MPDClient.MaintainMPDState: %v", err)
			continue
		}
		log.Debugf("MPDClient.MaintainMPDState curSongData.Title: %s", curSongData.Title)

		log.Debugf("Queue items:")
		entries := m.GetPlayQueue()
		allKeys := []int{}
		for key, _ := range m.GetPlayQueue() {
			allKeys = append(allKeys, key)
		}

		sort.Ints(allKeys)

		for _, idx := range allKeys {
			entry := entries[idx]
			log.Debugf("#%d: %v", idx, entry.Filename[:len(entry.Filename)-16])
		}

		curSongData.RequestQueue = m.GetPlayQueue()

		tDuration := time.Since(tStart)
		log.Debugf("MPDClient.MaintainMPDState: updated state in %v", tDuration)

		m.np = curSongData
	}
}

func (m *MPDClient) NowPlaying() NowPlayingData {
	m.np.ImageUrl = m.imageUrl

	return m.np
}

func (m *MPDClient) Next() NowPlayingData {
	if m.queue.Size() > 0 {
		entries := m.GetPlayQueue().ToMap()
		entry := entries[0]
		m.conn.Play(entry.Pos)
	} else {
		m.conn.Next()
	}
	return m.NowPlaying()
}

func (m *MPDClient) Play() NowPlayingData {
	m.Shuffle()

	m.conn.Play(1)
	return m.NowPlaying()
}

func (m *MPDClient) PlayPos(pos int) NowPlayingData {
	m.conn.Play(pos)
	return m.NowPlaying()
}

func (m *MPDClient) Shuffle() {
	m.conn.Shuffle(-1, -1)
}

func (m *MPDClient) UpdateDB(fname string) error {
	_, err := m.conn.Update(fname)
	time.Sleep(1 * time.Second)
	m.Add(fname)
	return err
}

func (m *MPDClient) Add(fileName string) {
	log.Infof("Adding %s to playlist", fileName)
	m.conn.Add(fileName)
}

func (m *MPDClient) PrioId(id, prio int) error {
	mpdPrio := 9 - prio
	err := m.conn.PrioId(mpdPrio, id)
	if err != nil {
		err = fmt.Errorf("MPDClient.UpdateQueueIDs: failed to set prio: %v", err)
		log.Warningf("%v", err)
		return err
	}

	return nil
}

func (m *MPDClient) GetTitle(pos int) (string, error) {
	info, err := m.conn.PlaylistInfo(pos, -1)
	if err != nil {
		return "", fmt.Errorf("MPDClient.GetTitle m.conn.PlaylistInfo: %v", err)
	}

	return info[0]["file"], nil
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
				return -1, fmt.Errorf("MPDClient.Search strconv.Atoi: %v",
					err)
			}
			return pos, nil
		}
	}

	return -1, fmt.Errorf("MPDClient.Search: failed to search mpd")
}

func (m *MPDClient) GetPlaylist() (Playlist, error) {
	playlist, err := m.conn.PlaylistInfo(-1, -1)
	if err != nil {
		return nil, fmt.Errorf("MPDClient.GetPlaylist m.conn.PlaylistInfo: %v", err)
	}

	entries := Playlist{}

	for _, entry := range playlist {
		fname := path.Base(entry["file"])
		if entries.HasFilename(entry["file"]) {
			continue
		}

		duration := 0
		if entry["Time"] != "" {
			duration, err = strconv.Atoi(entry["Time"])
			if err != nil {
				return nil, fmt.Errorf("MPDClient.GetPlaylist strconv.Atoi: %v", err)
			}
		}

		rating := -1
		if entry["Track"] != "" {
			rating, err = strconv.Atoi(entry["Track"])
			if err != nil {
				return nil, fmt.Errorf("MPDClient.GetPlaylist strconv.Atoi: %v", err)
			}
		}

		pos := -1
		if entry["Pos"] != "" {
			pos, err = strconv.Atoi(entry["Pos"])
			if err != nil {
				return nil, fmt.Errorf("MPDClient.GetPlaylist strconv.Atoi: %v", err)
			}
		}

		id := -1
		if entry["Id"] != "" {
			id, err = strconv.Atoi(entry["Id"])
			if err != nil {
				return nil, fmt.Errorf("MPDClient.GetPlaylist strconv.Atoi: %v", err)
			}
		}

		prio := 0
		curPrio, ok := entry["Prio"]
		if ok {
			prio, err = strconv.Atoi(curPrio)
			if err != nil {
				return nil, fmt.Errorf("MPDClient.GetPlaylist: failed to convert string to int")
			}
			log.Debugf("MPDClient.GetPlaylist: %s has prio %d", fname[:len(fname)-16], prio)
		}

		item := &PlaylistEntry{
			Duration: duration,
			Title:    fname[:len(fname)-16],
			Filename: fname,
			Rating:   rating,
			Pos:      pos,
			Id:       id,
			Prio:     prio,
		}

		entries[fname] = item
	}

	return entries, nil
}

func (m *MPDClient) TracksForArtist(artist string) (Playlist, error) {
	result, err := m.conn.Search("artist", artist)
	if err != nil {
		return nil, fmt.Errorf("MPDClient.TracksForArtist m.conn.Search: %v", err)
	}

	entries := Playlist{}

	for _, entry := range result {
		duration, err := strconv.Atoi(entry["Time"])
		if err != nil {
			return nil, fmt.Errorf("MPDClient.TracksForArtist strconv.Atoi: %v", err)
		}

		rating, err := strconv.Atoi(entry["Track"])
		if err != nil {
			return nil, fmt.Errorf("MPDClient.TracksForArtist strconv.Atoi: %v", err)
		}

		item := &PlaylistEntry{
			Duration: duration,
			Artist:   entry["Artist"],
			Title:    entry["Title"],
			Filename: entry["file"],
			Rating:   rating,
		}

		entries[entry["file"]] = item
	}

	return entries, nil
}

func (m *MPDClient) PreLoadPlayQueue() error {
	playlist, err := m.GetPlaylist()
	if err != nil {
		return fmt.Errorf("MPDClient.PreLoadPlayQueue: %v", err)
	}

	preloadQueue := PlayQueueEntries{}

	for _, entry := range playlist {
		if entry.Prio > 0 {
			preloadQueue[entry.Id] = entry
		}
	}

	m.queue.entries = preloadQueue
	m.np.RequestQueue = preloadQueue

	log.Debugf("MPDClient.PreLoadPlayQueue: preloaded %d items from mpd playlist", len(preloadQueue))

	return nil
}
