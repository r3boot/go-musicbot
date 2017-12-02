package ircclient

import (
	"crypto/tls"
	"fmt"
	"go-ircevent"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
	"github.com/r3boot/go-musicbot/lib/logger"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/ytclient"
)

var log *logger.Logger

func NewIrcClient(l *logger.Logger, config *config.MusicBotConfig, mpdClient *mpdclient.MPDClient, ytClient *youtubeclient.YoutubeClient, id3 *id3tags.ID3Tags) *IrcClient {
	log = l

	client := &IrcClient{
		config:    config,
		mpdClient: mpdClient,
		ytClient:  ytClient,
		id3:       id3,
		Online:    make(map[string]bool),
	}

	client.conn = irc.IRC(config.IRC.Nickname, config.IRC.Nickname)
	client.conn.VerboseCallbackHandler = false
	client.conn.Debug = false
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
