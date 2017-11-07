package ircclient

import (
	"crypto/tls"
	"fmt"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
	"github.com/thoj/go-ircevent"
)

func NewIrcClient(config *config.MusicBotConfig, mpdClient *mpdclient.MPDClient, ytClient *youtubeclient.YoutubeClient, mp3Library *mp3lib.MP3Library) *IrcClient {
	client := &IrcClient{
		config:     config,
		mpdClient:  mpdClient,
		ytClient:   ytClient,
		mp3Library: mp3Library,
		Online:     make(map[string]bool),
	}

	client.conn = irc.IRC(config.IRC.Nickname, config.IRC.Nickname)
	client.conn.VerboseCallbackHandler = config.IRC.Debug
	client.conn.Debug = config.IRC.Debug
	client.conn.UseTLS = config.IRC.UseTLS
	client.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	client.initCallbacks()

	return client
}

func (c *IrcClient) Run() error {
	server := fmt.Sprintf("%s:%d", c.config.IRC.Server, c.config.IRC.Port)

	err := c.conn.Connect(server)
	if err != nil {
		return fmt.Errorf("Err %s", err)
	}

	go c.CheckIfSjaakIsOnline()

	c.conn.Loop()

	return nil
}
