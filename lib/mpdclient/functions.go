package mpdclient

import (
	"fmt"
	"time"

	"strconv"

	"gompd/mpd"
	"sort"

	"sync"

	"github.com/r3boot/go-musicbot/lib/albumart"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
)

type NowPlayingData struct {
	Title        string
	Duration     float64
	Elapsed      float64
	Remaining    float64
	Rating       int
	ImageUrl     string
	Filename     string
	Id           int
	RequestQueue PlayQueueEntries
}

type PlayQueueEntries map[int]*PlaylistEntry

type PlayQueue struct {
	max     int
	length  int
	conn    *mpd.Client
	entries PlayQueueEntries
	mutex   sync.RWMutex
}

type Artists []string

type MPDClient struct {
	baseDir  string
	address  string
	password string
	id3      *id3tags.ID3Tags
	art      *albumart.AlbumArt
	Config   *config.MusicBotConfig
	conn     *mpd.Client
	np       NowPlayingData
	curFile  string
	imageUrl string
	queue    *PlayQueue
}

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
