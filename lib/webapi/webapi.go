package webapi

import (
	"fmt"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
	"github.com/r3boot/go-musicbot/lib/logger"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
)

var log *logger.Logger

func NewWebAPI(l *logger.Logger, cfg *config.MusicBotConfig, mpd *mpdclient.MPDClient, id3 *id3tags.ID3Tags, yt *youtubeclient.YoutubeClient) *WebAPI {
	log = l

	address := fmt.Sprintf("%s:%s", cfg.Api.Address, cfg.Api.Port)
	assets := cfg.Api.Assets

	api := &WebAPI{
		address:   address,
		assets:    assets,
		mpdClient: mpd,
		id3Tags:   id3,
		youtube:   yt,
		Config:    cfg,
	}

	go api.UpdatePlaylist()

	return api
}
