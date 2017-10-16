package ircclient

import (
	"regexp"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
	"github.com/thoj/go-ircevent"
)

const (
	CMD_DJPLUS   string = "dj+"
	CMD_START    string = "start"
	CMD_NEXT     string = "next"
	CMD_PLAYING  string = "np"
	CMD_RADIO    string = "radio"
	CMD_BOO      string = "boo"
	CMD_TUNE     string = "tune"
	CMD_PLAYLIST string = "djlist"
	CMD_PLAY     string = "play"
)

var (
	RE_CMD       = regexp.MustCompile("^(\\![a-z\\+\\-]{2,6})")
	RE_DJHANDLER = regexp.MustCompile("(\\!dj\\+) ([a-zA-Z0-9_-]{11})")
	RE_DJLIST    = regexp.MustCompile("(\\!djlist) (https://www.youtube.com/watch.*list=.*)")
	RE_SEARCH    = regexp.MustCompile("(\\!play) ([a-zA-Z0-9_\\-\\.\\ ]+)$")
)

type IrcClient struct {
	config     *config.MusicBotConfig
	conn       *irc.Connection
	mpdClient  *mpdclient.MPDClient
	ytClient   *youtubeclient.YoutubeClient
	mp3Library *mp3lib.MP3Library
}
