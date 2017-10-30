package webapi

import (
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
)

type TemplateData struct {
	Title  string
	Stream string
}

type WebApi struct {
	config *config.MusicBotConfig
	mpd    *mpdclient.MPDClient
	mp3    *mp3lib.MP3Library
	yt     *youtubeclient.YoutubeClient
}

type ClientRequest struct {
	Operation string
}

type NowPlayingResp struct {
	Title    string
	Duration string
	Rating   int
	Pkt      string
}
