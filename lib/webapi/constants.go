package webapi

import (
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
)

const (
	TF_CLF = "02/Jan/2006:15:04:05 -0700"

	WS_GET_PLAYLIST = 0
	WS_GET_ARTISTS  = 1
	WS_NEXT         = 2
	WS_BOO          = 3
	WS_TUNE         = 4
	WS_NOWPLAYING   = 5
	WS_REQUEST      = 6
)

type NowPlayingData struct {
	Artist   string `json:"artist"`
	Title    string `json:"title"`
	Name     string `json:"name"`
	Duration int    `json:"duration"`
	Rating   int    `json:"rating"`
	ImageUrl string `json:"image_url"`
}

type IndexTemplateData struct {
	Title      string
	NowPlaying string
	Mp3Stream  string
}

type WebResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type WSClientRequest struct {
	Id        int         `json:"i"`
	Operation int         `json:"o"`
	Data      interface{} `json:"d"`
}

type WSServerResponse struct {
	ClientId  int         `json:"i"`
	Operation int         `json:"o"`
	Status    bool        `json:"s"`
	Message   string      `json:"m"`
	Data      interface{} `json:"d"`
}

type WebAPI struct {
	address   string
	assets    string
	mpdClient *mpdclient.MPDClient
	id3Tags   *id3tags.ID3Tags
	youtube   *youtubeclient.YoutubeClient
	Config    *config.MusicBotConfig
	Playlist  mpdclient.Playlist
	Artists   mpdclient.Artists
}
