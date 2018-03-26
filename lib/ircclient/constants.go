package ircclient

import (
	"regexp"

	"go-ircevent"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
)

const (
	CMD_DJPLUS        = "!dj+"
	CMD_START         = "!start"
	CMD_NEXT          = "!next"
	CMD_PLAYING       = "!np"
	CMD_RADIO         = "!radio"
	CMD_BOO           = "!boo"
	CMD_TUNE          = "!tune"
	CMD_CH00N         = "!ch00n"
	CMD_REQUEST       = "!request"
	CMD_QUEUE         = "!queue"
	CMD_HELP          = "!help"
	NICK_SJAAK        = "Sjaak"
	NICK_FLUNK        = "flunk"
	RESPONSE_HELP     = "Available commands: dj+ <yt vid id>, request <query>, query, np, next, tune, boo, start, radio"
	UNKNOWN_SUBMITTER = "musicbot"
)

var (
	RE_CMD       = regexp.MustCompile("^(\\![a-z0-9\\+\\-]{2,8})")
	RE_DJHANDLER = regexp.MustCompile("(\\!dj\\+) ([a-zA-Z0-9_-]{11})")
	RE_DJLIST    = regexp.MustCompile("(\\!djlist) (https://www.youtube.com/watch.*list=.*)")
	RE_SEARCH    = regexp.MustCompile("(\\!request) ([a-zA-Z0-9_\\-\\.\\ ]+)$")
)

type IrcClient struct {
	config    *config.MusicBotConfig
	conn      *irc.Connection
	mpdClient *mpdclient.MPDClient
	ytClient  *youtubeclient.YoutubeClient
	id3       *id3tags.ID3Tags
	Online    map[string]bool
}
