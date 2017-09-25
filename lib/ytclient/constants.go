package youtubeclient

import (
	"sync"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
)

const (
	MAX_DOWNLOAD_QUEUE_SIZE int = 128
)

type YoutubeClient struct {
	seenFileMutex sync.RWMutex
	downloadMutex sync.RWMutex
	config        *config.MusicBotConfig
	mpdClient     *mpdclient.MPDClient
	musicDir      string
	DownloadChan  chan string
}
