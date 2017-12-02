package mpdclient

import (
	"gompd/mpd"
	"sync"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
)

const (
	MAX_QUEUE_ITEMS int = 8
)

type NowPlayingData struct {
	Title     string
	Duration  float64
	Elapsed   float64
	Remaining float64
	Rating    int
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

type MPDClient struct {
	config  *config.MusicBotConfig
	id3     *id3tags.ID3Tags
	address string
	conn    *mpd.Client
	np      NowPlayingData
	queue   *RequestQueue
}
