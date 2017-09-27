package youtubeclient

import (
	"sync"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
)

func NewYoutubeClient(config *config.MusicBotConfig, mpdclient *mpdclient.MPDClient, mp3Library *mp3lib.MP3Library, musicDir string) *YoutubeClient {
	yt := &YoutubeClient{
		seenFileMutex: sync.RWMutex{},
		downloadMutex: sync.RWMutex{},
		config:        config,
		mpdClient:     mpdclient,
		mp3Library:    mp3Library,
		musicDir:      musicDir,
		DownloadChan:  make(chan string, MAX_DOWNLOAD_QUEUE_SIZE),
	}

	go yt.DownloadSerializer()

	return yt
}
