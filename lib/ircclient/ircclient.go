package ircclient

import (
	"crypto/tls"
	"fmt"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/logger"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
	"go-ircevent"
)

var log *logger.Logger

func NewIrcClient(l *logger.Logger, config *config.MusicBotConfig, mpdClient *mpdclient.MPDClient, ytClient *youtubeclient.YoutubeClient, mp3Library *mp3lib.MP3Library) *IrcClient {
	log = l

	client := &IrcClient{
		config:     config,
		mpdClient:  mpdClient,
		ytClient:   ytClient,
		mp3Library: mp3Library,
		Online:     make(map[string]bool),
	}

	client.conn = irc.IRC(config.IRC.Nickname, config.IRC.Nickname)
	client.conn.VerboseCallbackHandler = config.App.Debug
	client.conn.Debug = config.App.Debug
	client.conn.UseTLS = config.IRC.UseTLS
	client.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	client.initCallbacks()

	return client
}

func (c *IrcClient) Run() error {
	server := fmt.Sprintf("%s:%d", c.config.IRC.Server, c.config.IRC.Port)

	err := c.conn.Connect(server)
	if err != nil {
		return fmt.Errorf("IrcClient.Run c.conn.Connect: %v", err)
	}

	go c.CheckIfSjaakIsOnline()

	c.conn.Loop()

	return nil
}
