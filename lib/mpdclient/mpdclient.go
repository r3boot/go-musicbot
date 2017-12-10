package mpdclient

import (
	"fmt"

	"github.com/r3boot/go-musicbot/lib/albumart"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
	"github.com/r3boot/go-musicbot/lib/logger"
	"sync"
)

var log *logger.Logger

func NewMPDClient(l *logger.Logger, cfg *config.MusicBotConfig, id3 *id3tags.ID3Tags, art *albumart.AlbumArt, baseDir string) (*MPDClient, error) {
	log = l

	address := fmt.Sprintf("%s:%d", cfg.MPD.Address, cfg.MPD.Port)

	client := &MPDClient{
		art:      art,
		id3:      id3,
		Config:   cfg,
		baseDir:  baseDir,
		address:  address,
		password: cfg.MPD.Password,
		np:       NowPlayingData{},
		queue: &PlayQueue{
			entries: make(PlayQueueEntries, MAX_QUEUE_ITEMS),
			mutex:   sync.RWMutex{},
		},
	}

	err := client.Connect()
	if err != nil {
		return nil, fmt.Errorf("NewMPDClient: %v", err)
	}

	go client.MaintainMPDState()

	return client, nil
}
