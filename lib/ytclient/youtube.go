package youtubeclient

import (
	"sync"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/logger"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
)

var log *logger.Logger

func NewYoutubeClient(l *logger.Logger, config *config.MusicBotConfig, mpdclient *mpdclient.MPDClient, mp3Library *mp3lib.MP3Library, musicDir string) *YoutubeClient {
	log = l
	yt := &YoutubeClient{
		seenFileMutex: sync.RWMutex{},
		downloadMutex: sync.RWMutex{},
		config:        config,
		mpdClient:     mpdclient,
		mp3Library:    mp3Library,
		MusicDir:      musicDir,
		DownloadChan:  make(chan string, MAX_DOWNLOAD_QUEUE_SIZE),
		PlaylistChan:  make(chan string, MAX_DOWNLOAD_QUEUE_SIZE),
	}

	go yt.DownloadSerializer()
	go yt.PlaylistSerializer()

	return yt
}
