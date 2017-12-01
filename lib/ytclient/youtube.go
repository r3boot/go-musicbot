package youtubeclient

import (
	"sync"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
	"github.com/r3boot/go-musicbot/lib/logger"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
)

var log *logger.Logger

func NewYoutubeClient(l *logger.Logger, config *config.MusicBotConfig, mpdclient *mpdclient.MPDClient, id3 *id3tags.ID3Tags, musicDir string) *YoutubeClient {
	log = l
	yt := &YoutubeClient{
		seenFileMutex: sync.RWMutex{},
		downloadMutex: sync.RWMutex{},
		mpdMutex:      sync.RWMutex{},
		config:        config,
		mpdClient:     mpdclient,
		id3:           id3,
		MusicDir:      musicDir,
		DownloadChan:  make(chan string, MAX_DOWNLOAD_QUEUE_SIZE),
		PlaylistChan:  make(chan string, MAX_DOWNLOAD_QUEUE_SIZE),
	}

	// Start download workers
	num := 0
	for id := 1; id <= yt.config.Youtube.NumWorkers; id++ {
		go yt.DownloadWorker(id, yt.DownloadChan)
		num += 1
	}

	log.Debugf("NewYoutubeClient: Started %d download workers", num)

	go yt.PlaylistSerializer()

	return yt
}
