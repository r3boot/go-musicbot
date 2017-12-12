package mpdclient

import (
	"encoding/json"
	"fmt"
)

type PlaylistEntry struct {
	Artist   string `json:"artist"`
	Title    string `json:"title"`
	Name     string `json:"name"`
	Rating   int    `json:"rating"`
	Filename string `json:"filename"`
	Duration int    `json:"duration"`
	Pos      int    `json:"pos"`
	Id       int    `json:"id"`
	QPrio    int    `json:"prio"`
	prio     int
}

type Playlist map[string]*PlaylistEntry

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
