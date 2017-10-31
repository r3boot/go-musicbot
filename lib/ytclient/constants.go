package youtubeclient

import (
	"sync"

	"regexp"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
)

const (
	MAX_DOWNLOAD_QUEUE_SIZE int = 128
)

var (
	RE_DESTINATION = regexp.MustCompile("\\[ffmpeg\\] Destination: (.*)-([a-zA-Z0-9_-]{11}).mp3")
)

type YoutubeClient struct {
	seenFileMutex sync.RWMutex
	downloadMutex sync.RWMutex
	config        *config.MusicBotConfig
	mpdClient     *mpdclient.MPDClient
	mp3Library    *mp3lib.MP3Library
	MusicDir      string
	DownloadChan  chan string
	PlaylistChan  chan string
}
