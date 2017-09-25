package youtubeclient

import (
	"sync"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
)

func NewYoutubeClient(config *config.MusicBotConfig, mpdclient *mpdclient.MPDClient, musicDir string) *YoutubeClient {
	yt := &YoutubeClient{
		seenFileMutex: sync.RWMutex{},
		downloadMutex: sync.RWMutex{},
		config:        config,
		mpdClient:     mpdclient,
		musicDir:      musicDir,
		DownloadChan:  make(chan string, MAX_DOWNLOAD_QUEUE_SIZE),
	}

	return yt
}
