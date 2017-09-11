package youtubeclient

import (
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"sync"
)

type YoutubeClient struct {
	seenFileMutex sync.RWMutex
	downloadMutex sync.RWMutex
	config        *config.MusicBotConfig
	mpdClient     *mpdclient.MPDClient
	musicDir      string
}
