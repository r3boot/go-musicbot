package mpdclient

import (
	"sync"

	"gompd/mpd"

	"github.com/r3boot/go-musicbot/lib/albumart"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
)

const (
	MAX_QUEUE_ITEMS = 8
)

type NowPlayingData struct {
	Title        string
	Duration     float64
	Elapsed      float64
	Remaining    float64
	Rating       int
	ImageUrl     string
	Filename     string
	RequestQueue PlayQueueEntries
}

type RequestQueueItem struct {
	Title string
	Pos   int
}

type RequestQueue struct {
	entries []*RequestQueueItem
	size    int
	count   int
	mutex   sync.RWMutex
}

type PlaylistEntry struct {
	Artist   string `json:"artist"`
	Title    string `json:"title"`
	Rating   int    `json:"rating"`
	Filename string `json:"filename"`
	Duration int    `json:"duration"`
	Pos      int    `json:"pos"`
	Id       int    `json:"id"`
	Prio     int    `json:"prio"`
}

type Playlist map[string]*PlaylistEntry
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
