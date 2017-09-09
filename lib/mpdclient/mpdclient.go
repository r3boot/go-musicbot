package mpdclient

import (
	"fmt"

	"github.com/r3boot/go-musicbot/lib/config"
)

func NewMPDClient(config *config.MusicBotConfig) (*MPDClient, error) {
	client := &MPDClient{
		config:  config,
		address: fmt.Sprintf("%s:%d", config.MPD.Address, config.MPD.Port),
	}

	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to initialized client: %v", err)
	}

	go client.KeepAlive()

	return client, nil
}
