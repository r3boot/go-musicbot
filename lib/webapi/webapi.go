package webapi

import (
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
	"github.com/r3boot/go-musicbot/lib/logger"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
)

var log *logger.Logger

func NewWebApi(l *logger.Logger, cfg *config.MusicBotConfig, mpd *mpdclient.MPDClient, id3 *id3tags.ID3Tags, yt *youtubeclient.YoutubeClient) *WebApi {
	log = l

	api := &WebApi{
		config: cfg,
		mpd:    mpd,
		id3:    id3,
		yt:     yt,
	}

	go api.updateNowPlayingData()

	return api
}
