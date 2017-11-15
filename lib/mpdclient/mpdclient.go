package mpdclient

import (
	"fmt"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
)

func NewMPDClient(config *config.MusicBotConfig, mp3 *mp3lib.MP3Library) (*MPDClient, error) {
	client := &MPDClient{
		config:  config,
		mp3:     mp3,
		address: fmt.Sprintf("%s:%d", config.MPD.Address, config.MPD.Port),
		np:      NowPlayingData{},
		queue:   NewRequestQueue(MAX_QUEUE_ITEMS),
	}

	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to initialized client: %v", err)
	}

	go client.KeepAlive()
	go client.MaintainMPDState()

	return client, nil
}
