package youtubeclient

import (
	"sync"

	"regexp"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
)

const (
	maxDownloadQueueSize = 128
	MaxSongLength        = 1800 // In seconds
)

var (
	reDestination = regexp.MustCompile("\\[ffmpeg\\] Destination: (.*)-([a-zA-Z0-9_-]{11}).mp3")
	reSongLength  = regexp.MustCompile("\"length_seconds\":\"([0-9]+)\"")
)

type YoutubeClient struct {
	seenFileMutex sync.RWMutex
	downloadMutex sync.RWMutex
	mpdMutex      sync.RWMutex
	config        *config.MusicBotConfig
	mpdClient     *mpdclient.MPDClient
	id3           *id3tags.ID3Tags
	MusicDir      string
	DownloadChan  chan DownloadMeta
	PlaylistChan  chan string
}

type DownloadMeta struct {
	Yid       string
	Nickname  string
	IsRequest bool
}
