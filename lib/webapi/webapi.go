package webapi

import (
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
)

func NewWebApi(cfg *config.MusicBotConfig, mpd *mpdclient.MPDClient, mp3 *mp3lib.MP3Library, yt *youtubeclient.YoutubeClient) *WebApi {
	return &WebApi{
		config: cfg,
		mpd:    mpd,
		mp3:    mp3,
		yt:     yt,
	}
}
