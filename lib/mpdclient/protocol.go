package mpdclient

import (
	"fmt"
	"path"
	"strconv"
	"time"
)

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
			log.Debugf("MPDClient.GetPlaylist: %s has prio %d, id %d", fname[:len(fname)-16], prio, id)
		}

		item := &PlaylistEntry{
			Duration: duration,
			Title:    fname[:len(fname)-16],
			Name:     fname[:len(fname)-16],
			Filename: fname,
			Rating:   rating,
			Pos:      pos,
			Id:       id,
			prio:     prio,
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
