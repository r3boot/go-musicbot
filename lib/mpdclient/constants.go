package mpdclient

import (
	"github.com/fhs/gompd/mpd"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
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
	Count   int
}

type MPDClient struct {
	config  *config.MusicBotConfig
	mp3     *mp3lib.MP3Library
	address string
	conn    *mpd.Client
	np      NowPlayingData
	queue   *RequestQueue
}
