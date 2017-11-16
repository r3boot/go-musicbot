package webapi

import (
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/logger"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
)

var log *logger.Logger

func NewWebApi(l *logger.Logger, cfg *config.MusicBotConfig, mpd *mpdclient.MPDClient, mp3 *mp3lib.MP3Library, yt *youtubeclient.YoutubeClient) *WebApi {
	log = l

	api := &WebApi{
		config: cfg,
		mpd:    mpd,
		mp3:    mp3,
		yt:     yt,
	}

	go api.updateNowPlayingData()

	return api
}
