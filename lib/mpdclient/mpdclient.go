package mpdclient

import (
	"fmt"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
	"github.com/r3boot/go-musicbot/lib/logger"
)

var log *logger.Logger

func NewMPDClient(l *logger.Logger, config *config.MusicBotConfig, id3 *id3tags.ID3Tags) (*MPDClient, error) {
	log = l

	client := &MPDClient{
		config:  config,
		id3:     id3,
		address: fmt.Sprintf("%s:%d", config.MPD.Address, config.MPD.Port),
		np:      NowPlayingData{},
		queue:   NewRequestQueue(MAX_QUEUE_ITEMS),
	}

	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("NewMPDClient client.Connect: %v", err)
	}

	go client.KeepAlive()
	go client.MaintainMPDState()

	return client, nil
}
