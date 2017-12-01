package youtubeclient

import (
	"sync"

	"regexp"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
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
	mpdMutex      sync.RWMutex
	config        *config.MusicBotConfig
	mpdClient     *mpdclient.MPDClient
	id3           *id3tags.ID3Tags
	MusicDir      string
	DownloadChan  chan string
	PlaylistChan  chan string
}
