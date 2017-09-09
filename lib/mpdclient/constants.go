package mpdclient

import (
	"github.com/fhs/gompd/mpd"
	"github.com/r3boot/go-musicbot/lib/config"
)

type MPDClient struct {
	config  *config.MusicBotConfig
	address string
	conn    *mpd.Client
}
