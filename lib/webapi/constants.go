package webapi

import (
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
)

const (
	MAX_PLAYLIST_LENGTH int    = 8192
	TF_CLF              string = "02/Jan/2006:15:04:05 -0700"
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

type SearchRequest struct {
	Operation string
	Query     string
}

type CachedData struct {
	Title    string
	Duration string
	Rating   int
	Playlist []string
}

type AutoCompleteRequest struct {
	Query string `json:"query"`
}

type AutoCompleteResponse struct {
	Query       string   `json:"query"`
	Suggestions []string `json:"suggestions"`
}

type NowPlaying struct {
	Title    string
	Duration string
	Rating   int
}

type NowPlayingResp struct {
	Data NowPlaying
	Pkt  string
}

type GetQueueRespData struct {
	Entries map[int]string
	Size    int
}

type GetQueueResp struct {
	Data GetQueueRespData
	Pkt  string
}

type GetFilesResp struct {
	Data []string
	Pkt  string
}
