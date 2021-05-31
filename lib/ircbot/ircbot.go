package ircbot

import (
	"crypto/tls"
	"fmt"
	"github.com/go-openapi/runtime"
	"github.com/r3boot/go-musicbot/lib/apiclient"
	"github.com/r3boot/go-musicbot/lib/apiclient/operations"
	"github.com/r3boot/go-musicbot/lib/config"
	irc "github.com/thoj/go-ircevent"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const (
	// TODO: Make commande dependent on control character
	cmdDjPlus        = "!dj+"
	cmdNext          = "!next"
	cmdNowPlaying    = "!np"
	cmdRadio         = "!radio"
	cmdBoo           = "!boo"
	cmdTune          = "!tune"
	cmdCh00n         = "!ch00n"
	cmdRequest       = "!request"
	cmdQueue         = "!queue"
	cmdSearch        = "!search"
	cmdHelp          = "!help"
	responseHelp     = "Available commands: dj+ <yt vid id>, request <query>, search <query>, queue, np, next, tune, boo, ch00n, radio"
	unknownSubmitter = "musicbot"
	maxQueryLength   = 256
)

type IrcBotParams struct {
	Nickname  string
	Server    string
	Port      int
	Channel   string
	Tls       bool
	VerifyTls bool
}

type IrcBot struct {
	config *config.Config
	conn   *irc.Connection
	api    *apiclient.Musicbot
	token  runtime.ClientAuthInfoWriter
	params *IrcBotParams
}

var (
	// TODO: Make these dependent on the control character
	reCmd       = regexp.MustCompile("^(\\![a-z0-9\\+\\-]{2,8})")
	reDjHandler = regexp.MustCompile("(\\!dj\\+) ([a-zA-Z0-9_-]{11})")
	reRequest   = regexp.MustCompile("(\\!request) ([a-zA-Z0-9_\\-\\.\\ ]+)$")
	reSearch    = regexp.MustCompile("(\\!search) ([a-zA-Z0-9_\\-\\.\\ ]+)$")
)

func NewIrcBot(config *config.Config, client *apiclient.Musicbot, token runtime.ClientAuthInfoWriter, params *IrcBotParams) (*IrcBot, error) {
	bot := &IrcBot{
		config: config,
		api:    client,
		token:  token,
		params: params,
	}

	bot.conn = irc.IRC(params.Nickname, params.Nickname)
	bot.conn.VerboseCallbackHandler = false
	bot.conn.Debug = false
	bot.conn.UseTLS = params.Tls
	bot.conn.TLSConfig = &tls.Config{InsecureSkipVerify: !params.VerifyTls}

	bot.conn.AddCallback("001", func(e *irc.Event) { bot.conn.Join(bot.params.Channel) })
	bot.conn.AddCallback("PRIVMSG", bot.ParsePrivmsg)

	return bot, nil
}

func (c *IrcBot) Run() error {
	server := fmt.Sprintf("%s:%d", c.params.Server, c.params.Port)

	err := c.conn.Connect(server)
	if err != nil {
		return fmt.Errorf("IrcBot.Run c.conn.Connect: %v", err)
	}

	c.conn.Loop()

	return nil
}

func (c *IrcBot) randomRadioReply() string {
	n := rand.Int() % len(c.config.IrcBot.RadioReplies)
	return c.config.IrcBot.RadioReplies[n]
}

func (c *IrcBot) randomCh00nReply() string {
	n := rand.Int() % len(c.config.IrcBot.Ch00nReplies)
	return c.config.IrcBot.Ch00nReplies[n]
}

func (c *IrcBot) ParsePrivmsg(e *irc.Event) {
	if len(e.Arguments) != 2 {
		return
	}

	channel := e.Arguments[0]
	line := e.Arguments[1]

	cmdResult := reCmd.FindAllStringSubmatch(line, -1)
	if len(cmdResult) != 1 {
		return
	}

	command := cmdResult[0][1]
	user := e.User
	if len(user) > 0 && user[0] == '~' {
		user = user[1:]
	}

	fmt.Printf("IrcBot.ParsePrivmsg: Got command %s", command)

	switch command {
	case cmdHelp:
		c.HandleHelp(channel)
	case cmdDjPlus:
		c.HandleDownload(channel, line, e.Nick)
	case cmdNext:
		c.HandleNext(channel, line, user)
	case cmdNowPlaying:
		c.HandleNowPlaying(channel, line)
	case cmdRadio:
		c.HandleRadioUrl(channel, line)
	case cmdBoo:
		c.HandleDecreaseRating(channel, line, user)
	case cmdTune:
		c.HandleIncreaseRating(channel, line, user)
	case cmdCh00n:
		c.HandleCh00n(channel, line, user)
	case cmdRequest:
		c.HandleRequest(channel, line, user)
	case cmdSearch:
		c.HandleSearch(channel, line, user)
	case cmdQueue:
		c.HandleQueue(channel, line, user)
	default:
		fmt.Printf("IrcBot.ParsePrivmsg: Invalid command received: %s", command)
	}
}

func (c *IrcBot) HandleHelp(channel string) {
	c.conn.Privmsg(channel, responseHelp)
}

func (c *IrcBot) HandleDownload(channel, line, user string) {
	result := reDjHandler.FindAllStringSubmatch(line, -1)
	response := "Undefined"

	if len(result) == 1 {
		yid := result[0][2]
		params := operations.NewPostTrackDownloadParams()
		params.Body = operations.PostTrackDownloadBody{
			Yid:       &yid,
			Submitter: &user,
		}

		resp, err := c.api.Operations.PostTrackDownload(params, c.token)
		if err != nil {
			fmt.Printf("IrcBot.HandleDownload: Failed to download track: %v", err)
			errmsg := err.Error()
			if strings.Contains(errmsg, "postTrackDownloadRequestEntityTooLarge") {
				response = fmt.Sprintf("Track is too long for stream")
				c.conn.Privmsg(channel, response)
			} else if strings.Contains(errmsg, "postTrackDownloadConflict") {
				response = fmt.Sprintf("Track is already downloaded")
				c.conn.Privmsg(channel, response)
			} else {
				response = fmt.Sprintf("Failed to download track")
				c.conn.Privmsg(channel, response)
			}
		}

		track := resp.GetPayload()
		fname := *track.Filename

		response = fmt.Sprintf("Added %s to the playlist", fname[:len(fname)-16])
		c.conn.Privmsg(channel, response)
	} else {
		fmt.Printf("IrcBot.HandleDownload: no results found")
		response = fmt.Sprintf("No yid found in message .. Anta BAKA??")
		c.conn.Privmsg(channel, response)
	}
}

func (c *IrcBot) HandleNext(channel, line, user string) {
	resp, err := c.api.Operations.GetPlayerNext(operations.NewGetPlayerNextParams(), c.token)
	if err != nil {
		fmt.Printf("IrcBot.HandleNext: failed to skip to next track: %v", err)
		response := fmt.Sprintf("Failed to skip track")
		c.conn.Privmsg(channel, response)
	}

	track := resp.GetPayload()
	fname := *track.Filename

	response := fmt.Sprintf("Now playing: %s", fname[:len(fname)-16])
	c.conn.Privmsg(channel, response)
}

func (c *IrcBot) HandleNowPlaying(channel, line string) {
	resp, err := c.api.Operations.GetPlayerNowplaying(operations.NewGetPlayerNowplayingParams(), c.token)
	if err != nil {
		fmt.Printf("Failed to fetch now playing: %v\n", err)
		response := "Failed to fetch now playing"
		c.conn.Privmsg(channel, response)
		return
	}

	track := resp.GetPayload()
	fname := *track.Filename

	durationFormat := fmt.Sprintf("%ds", int(*track.Duration))
	d, _ := time.ParseDuration(durationFormat)
	formattedDuration := d.String()

	response := fmt.Sprintf("Now playing: %s (duration: %s; rating: %d/10; submitter: %s)", fname[:len(fname)-16], formattedDuration, *track.Rating, *track.Submitter)
	c.conn.Privmsg(channel, response)
}

func (c *IrcBot) HandleRadioUrl(channel, line string) {
	response := fmt.Sprintf("%s Listen to %s", c.randomRadioReply(), c.config.IrcBot.StreamUrl)
	c.conn.Privmsg(channel, response)
}

func (c *IrcBot) HandleDecreaseRating(channel, line, user string) {
	resp, err := c.api.Operations.GetRatingDecrease(operations.NewGetRatingDecreaseParams(), c.token)
	if err != nil {
		fmt.Printf("Failed to decrease rating: %v\n", err)
		response := "Failed to decrease rating"
		c.conn.Privmsg(channel, response)
		return
	}

	track := resp.GetPayload()
	// TODO: Add detection for playlist removal

	fname := *track.Filename

	response := fmt.Sprintf("Rating for %s is %d/10 .. BOOO!!!!", fname[:len(fname)-16], *track.Rating)
	c.conn.Privmsg(channel, response)

	if *track.Submitter != "" && *track.Submitter != unknownSubmitter {
		submitterResponse := fmt.Sprintf("%s--", *track.Submitter)
		c.conn.Privmsg(channel, submitterResponse)
	}
}

func (c *IrcBot) HandleIncreaseRating(channel, line, user string) {
	resp, err := c.api.Operations.GetRatingIncrease(operations.NewGetRatingIncreaseParams(), c.token)
	if err != nil {
		fmt.Printf("Failed to increase rating: %v\n", err)
		response := "Failed to increase rating"
		c.conn.Privmsg(channel, response)
		return
	}

	track := resp.GetPayload()
	fname := *track.Filename

	response := fmt.Sprintf("Rating for %s is %d/10 .. Party on!!!!", fname[:len(fname)-16], *track.Rating)
	c.conn.Privmsg(channel, response)

	if *track.Submitter != "" && *track.Submitter != unknownSubmitter && *track.Submitter != user {
		submitterResponse := fmt.Sprintf("%s++", *track.Submitter)
		c.conn.Privmsg(channel, submitterResponse)
	}
}

func (c *IrcBot) HandleCh00n(channel, line, user string) {
	response := fmt.Sprintf("%s", c.randomCh00nReply())
	response = fmt.Sprintf(response, user)
	c.conn.Privmsg(channel, response)
}

func (c *IrcBot) HandleRequest(channel, line, user string) {
	result := reRequest.FindAllStringSubmatch(line, -1)
	response := "Undefined"

	if len(result) == 1 {
		if len(result[0][2]) > maxQueryLength {
			fmt.Printf("IrcBot.HandleRequest: query too large")
			response = fmt.Sprintf("Size of query too large")
			c.conn.Privmsg(channel, response)
			return
		}

		query := result[0][2]

		params := operations.NewPostTrackRequestParams()
		params.Request = operations.PostTrackRequestBody{
			Query:     &query,
			Submitter: &user,
		}

		response, err := c.api.Operations.PostTrackRequest(params, c.token)
		if err != nil {
			fmt.Printf("IrcBot.HandleRequest: failed to submit query: %v", err)
			msg := fmt.Sprintf("Failed to submit query")
			c.conn.Privmsg(channel, msg)
			return
		}

		track := response.GetPayload()
		fname := *track.Track.Filename

		fmt.Printf("Added %s to the queue", fname[:len(fname)-16])
		msg := fmt.Sprintf("Added %s to the queue", fname[:len(fname)-16])
		c.conn.Privmsg(channel, msg)
	} else {
		fmt.Printf("IrcBot.HandleQuery: No query found")
		response = fmt.Sprintf("Need a query to search .. stupid!")
		c.conn.Privmsg(channel, response)
	}
}

func (c *IrcBot) HandleSearch(channel, line, user string) {
	result := reSearch.FindAllStringSubmatch(line, -1)

	if len(result) == 1 {
		query := result[0][2]

		params := operations.NewPostTrackSearchParams()
		params.Request = operations.PostTrackSearchBody{
			Query:     &query,
			Submitter: &user,
		}

		response, err := c.api.Operations.PostTrackSearch(params, c.token)
		if err != nil {
			fmt.Printf("Failed to search: %v\n", err)
			msg := fmt.Sprintf("Search failed")
			c.conn.Privmsg(channel, msg)
			return
		}

		foundEntries := response.GetPayload()

		if len(foundEntries) == 0 {
			msg := fmt.Sprintf("Nothing found")
			c.conn.Privmsg(channel, msg)
			return
		}

		c.conn.Privmsg(channel, "Sending results via /query")
		c.conn.Privmsg(user, "Found the following tracks:")
		for i := 0; i < len(foundEntries); i++ {
			fname := *foundEntries[i].Filename
			msg := fmt.Sprintf("%d) %s\n", i, fname[:len(fname)-16])
			c.conn.Privmsg(user, msg)
		}
	}
}

func (c *IrcBot) HandleQueue(channel, line, user string) {
	response, err := c.api.Operations.GetPlayerQueue(operations.NewGetPlayerQueueParams(), c.token)
	if err != nil {
		fmt.Printf("Failed to fetch queue: %v\n", err)
		response := fmt.Sprintf("Failed to fetch queue")
		c.conn.Privmsg(channel, response)
		return
	}

	queueEntries := response.GetPayload()

	if len(queueEntries) == 0 {
		response := fmt.Sprintf("Queue is empty")
		c.conn.Privmsg(channel, response)
		return
	}

	c.conn.Privmsg(channel, "Current queue:")
	for i := 0; i < len(queueEntries); i++ {
		fname := *queueEntries[i].Filename
		response := fmt.Sprintf("%d) %s\n", i, fname[:len(fname)-16])
		c.conn.Privmsg(channel, response)
	}
}
