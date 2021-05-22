package ircbot

import (
	"crypto/tls"
	"fmt"
	"github.com/r3boot/go-musicbot/pkg/config"
	"github.com/r3boot/go-musicbot/pkg/downloader"
	"github.com/r3boot/go-musicbot/pkg/mpd"
	"github.com/r3boot/go-musicbot/pkg/mq"
	"github.com/sirupsen/logrus"
	"github.com/thoj/go-ircevent"
	"regexp"
	"sort"
)

const (
	ModuleName = "IrcBot"

	cmdDjPlus     = "!dj+"
	cmdStart      = "!start"
	cmdNext       = "!next"
	cmdNowPlaying = "!np"
	cmdRadio      = "!radio"
	cmdBoo        = "!boo"
	cmdTune       = "!tune"
	cmdCh00n      = "!ch00n"
	cmdRequest    = "!request"
	cmdQueue      = "!queue"
	cmdHelp       = "!help"

	responseHelp     = "Available commands: dj+ <yt vid id>, request <query>, query, np, next, tune, boo, start, radio"
	responseNoID     = "No ID found in privmsg .. Anta BAKA??"
	responseIDExists = "ID already downloaded"
)

type IrcBot struct {
	sendChan            chan mq.Message
	recvChan            chan mq.Message
	quit                chan bool
	log                 *logrus.Entry
	cfg                 *config.MusicBotConfig
	conn                *irc.Connection
	idList              []string
	nowPlaying          *mq.PlaylistEntry
	nowPlayingRequested bool
}

var (
	reCmd     = regexp.MustCompile("^(\\![a-z0-9\\+\\-]{2,8})")
	reDjPlus  = regexp.MustCompile("\\!dj\\+ ([a-zA-Z0-9_-]{11})")
	reRequest = regexp.MustCompile("\\!request ([a-zA-Z0-9_\\-\\.\\ ]+)$")
)

func NewIrcBot(cfg *config.MusicBotConfig) (*IrcBot, error) {
	ircBot := &IrcBot{
		sendChan: make(chan mq.Message, mq.MaxInFlight),
		recvChan: make(chan mq.Message, mq.MaxInFlight),
		quit:     make(chan bool),
		log: logrus.WithFields(logrus.Fields{
			"caller":  ModuleName,
			"channel": cfg.IrcBot.Channel,
		}),
		cfg: cfg,
	}

	ircBot.initClient()
	ircBot.initCallbacks()

	go ircBot.MessagePipe()

	return ircBot, nil
}

func (bot *IrcBot) GetSendChan() chan mq.Message {
	return bot.sendChan
}

func (bot *IrcBot) GetRecvChan() chan mq.Message {
	return bot.recvChan
}

func (bot *IrcBot) MessagePipe() {
	bot.log.Debug("Starting MessagePipe")
	for {
		select {
		case msg := <-bot.recvChan:
			{
				msgType := msg.GetMsgType()
				switch msgType {
				case mq.MsgNowPlaying:
					{
						entry := msg.GetContent().(*mq.PlaylistEntry)
						bot.setNowPlaying(entry)
					}
				case mq.MsgUpdateIDs:
					{
						updateIDsMsg := msg.GetContent().(*mq.UpdateIDsMessage)
						bot.setIDs(updateIDsMsg.IDs)
					}
				case mq.MsgSongTooLong:
					{
						stlMsg := msg.GetContent().(*mq.SongTooLongMessage)
						bot.handleSongTooLong(stlMsg.ID, stlMsg.Submitter)
					}
				case mq.MsgDownloadAccepted:
					{
						daMsg := msg.GetContent().(*mq.DownloadAcceptedMessage)
						bot.handleDownloadAccepted(daMsg.ID, daMsg.Submitter)
					}
				case mq.MsgDownloadCompleted:
					{
						dcMsg := msg.GetContent().(*mq.DownloadCompletedMessage)
						bot.handleDownloadCompleted(dcMsg.Filename, dcMsg.Submitter)
					}
				case mq.MsgSearchError:
					{
						seMsg := msg.GetContent().(*mq.SearchErrorMessage)
						bot.handleSearchFailed(seMsg.Error, seMsg.Submitter)
					}
				case mq.MsgQueueError:
					{
						qeMsg := msg.GetContent().(*mq.QueueErrorMessage)
						bot.handleQueueError(qeMsg.Error, qeMsg.Submitter)
					}
				case mq.MsgQueueResult:
					{
						qrMsg := msg.GetContent().(*mq.QueueResultMessage)
						bot.handleQueueResponse(qrMsg.Filename, qrMsg.Submitter, qrMsg.Prio)
					}
				case mq.MsgGetQueueResponse:
					{
						qrMsg := msg.GetContent().(*mq.GetQueueResponseMessage)
						bot.handleGetQueueResponse(qrMsg.Queue, qrMsg.Submitter)
					}
				}
			}
		case <-bot.quit:
			{
				return
			}
		}
	}
}

// IRC bot functionality
func (bot *IrcBot) initClient() {
	bot.conn = irc.IRC(bot.cfg.IrcBot.Nickname, bot.cfg.IrcBot.Nickname)
	bot.conn.VerboseCallbackHandler = false
	bot.conn.Debug = bot.cfg.IrcBot.Debug
	bot.conn.UseTLS = bot.cfg.IrcBot.UseTLS
	if bot.cfg.IrcBot.UseTLS {
		bot.conn.TLSConfig = &tls.Config{
			InsecureSkipVerify: bot.cfg.IrcBot.VerifyTLS,
		}
	}
}

func (bot *IrcBot) initCallbacks() {
	bot.conn.AddCallback("001", bot.handleJoinChan)
	bot.conn.AddCallback("PRIVMSG", bot.handlePrivMsg)
}

func (bot *IrcBot) Run() error {
	server := fmt.Sprintf("%s:%d", bot.cfg.IrcBot.Server, bot.cfg.IrcBot.Port)

	bot.log.WithFields(logrus.Fields{
		"server":   server,
		"tls":      bot.cfg.IrcBot.UseTLS,
		"nickname": bot.cfg.IrcBot.Nickname,
	}).Info("Connecting to IRC network")

	err := bot.conn.Connect(server)
	if err != nil {
		return fmt.Errorf("bot.conn.Connect: %v", err)
	}

	bot.log.WithFields(logrus.Fields{
		"server":   server,
		"nickname": bot.cfg.IrcBot.Nickname,
	}).Debug("Starting event loop")
	bot.conn.Loop()

	return nil
}

func (bot *IrcBot) handleJoinChan(e *irc.Event) {
	bot.log.Debug("Joining channel")
	bot.conn.Join(bot.cfg.IrcBot.Channel)
}

func (bot *IrcBot) handlePrivMsg(e *irc.Event) {
	if len(e.Arguments) != 2 {
		bot.log.Warn("PRIVMSG with invalid number of arguments")
		return
	}

	channel := e.Arguments[0]
	privmsg := e.Arguments[1]

	result := reCmd.FindAllStringSubmatch(privmsg, -1)
	if len(result) != 1 {
		bot.log.Warn("Unable to parse PRIVMSG for command")
		return
	}

	command := result[0][1]

	user := e.User
	if len(user) > 0 && user[0] == '~' {
		user = user[1:]
	}

	nickname := e.Nick

	bot.log.WithFields(logrus.Fields{
		"command":  command,
		"user":     user,
		"nickname": nickname,
	}).Debug("Received command")

	switch command {
	case cmdHelp:
		{
			bot.handleHelp(channel, privmsg)
		}
	case cmdDjPlus:
		{
			bot.handleDownload(channel, privmsg, nickname)
		}
	case cmdNowPlaying:
		{
			bot.handleNowPlaying()
		}
	case cmdRequest:
		{
			bot.handleRequest(channel, privmsg, nickname)
		}
	case cmdQueue:
		{
			bot.handleGetQueueRequest(nickname)
		}
	}
}

func (bot *IrcBot) handleHelp(channel, privmsg string) {
	bot.conn.Privmsg(channel, responseHelp)
}

// MessageQueue functionality
func (bot *IrcBot) setNowPlaying(entry *mq.PlaylistEntry) {
	bot.log.WithFields(logrus.Fields{
		"fname":     entry.Filename,
		"pos":       entry.Pos,
		"id":        entry.Id,
		"duration":  entry.Duration,
		"rating":    entry.Rating,
		"submitter": entry.Submitter,
	}).Debug("Storing NowPlaying info")

	bot.nowPlaying = entry

	if bot.nowPlayingRequested {
		name := entry.Filename[:len(entry.Filename)-16]
		submitter := "unknown"
		if entry.Submitter != "" {
			submitter = entry.Submitter
		}
		response := fmt.Sprintf("Now playing: %s (duration: %d; rating: %d/10; submitter: %s)", name, entry.Duration, entry.Rating, submitter)
		bot.conn.Privmsg(bot.cfg.IrcBot.Channel, response)
		bot.nowPlayingRequested = false
	}
}

func (bot *IrcBot) setIDs(idList []string) {
	bot.log.WithFields(logrus.Fields{
		"numtracks": len(idList),
	}).Debug("Updating ID list")

	bot.idList = idList
}

func (bot *IrcBot) HandlePlay() {
	bot.log.Debug("Sending play")
	playMsg := mq.NewPlayMessage(1)
	msg := mq.NewMessage(ModuleName, mq.MsgPlay, playMsg)
	bot.sendChan <- *msg
}

func (bot *IrcBot) HandleNext() {
	bot.log.Debug("Sending Next")
	nextMsg := mq.NewNextMessage()
	msg := mq.NewMessage(ModuleName, mq.MsgNext, nextMsg)
	bot.sendChan <- *msg
}

func (bot *IrcBot) HandleRewind() {
	bot.log.Debug("Sending Rewind")
	rewindMsg := mq.NewRewindMessage()
	msg := mq.NewMessage(ModuleName, mq.MsgRewind, rewindMsg)
	bot.sendChan <- *msg
}

func (bot *IrcBot) handleRequest(channel, privmsg, nickname string) {
	result := reRequest.FindAllStringSubmatch(privmsg, -1)
	if len(result) != 1 {
		bot.log.WithFields(logrus.Fields{
			"user":    nickname,
			"privmsg": privmsg,
		}).Debug("Unable to parse query")
	}

	query := result[0][1]

	bot.log.WithFields(logrus.Fields{
		"Query": query,
	})
	requestMsg := mq.NewRequestMessage(query, nickname)
	msg := mq.NewMessage(ModuleName, mq.MsgRequest, requestMsg)
	bot.sendChan <- *msg
}

func (bot *IrcBot) handleGetQueueRequest(nickname string) {
	rqMsg := mq.NewGetQueueRequestMessage(nickname)
	msg := mq.NewMessage(ModuleName, mq.MsgGetQueueRequest, rqMsg)
	bot.sendChan <- *msg
}

func (bot *IrcBot) handleGetQueueResponse(queue map[int]mq.PlaylistEntry, submitter string) {
	var ids []int

	if len(queue) == 0 {
		response := fmt.Sprintf("No tracks are queued currently")
		bot.conn.Privmsg(bot.cfg.IrcBot.Channel, response)
		return
	} else {
		response := fmt.Sprintf("%s: Sending queue results via privmsg", submitter)
		bot.conn.Privmsg(bot.cfg.IrcBot.Channel, response)
	}

	for k := range queue {
		ids = append(ids, k)
	}
	sort.Ints(ids)

	response := fmt.Sprintf("The following tracks are queued:")
	bot.conn.Privmsg(submitter, response)

	for _, prio := range ids {
		response := fmt.Sprintf("%d) %s", prio, queue[prio].Filename[:len(queue[prio].Filename)-16])
		bot.conn.Privmsg(submitter, response)
	}
}

func (bot *IrcBot) handleDownload(channel, privmsg, nickname string) {
	result := reDjPlus.FindAllStringSubmatch(privmsg, -1)

	if len(result) != 1 {
		bot.log.WithFields(logrus.Fields{
			"user":    nickname,
			"privmsg": privmsg,
		}).Debug("Unable to parse ID")
		bot.conn.Privmsg(channel, responseNoID)
		return
	}

	id := result[0][1]

	for _, entry := range bot.idList {
		if entry == id {
			bot.log.WithFields(logrus.Fields{
				"user": nickname,
				"id":   id,
			}).Warn("ID already downloaded")
			bot.conn.Privmsg(channel, responseIDExists)
			return
		}
	}

	site := downloader.SiteYoutube
	submitter := nickname

	bot.log.WithFields(logrus.Fields{
		"site":      downloader.SiteToString[site],
		"id":        id,
		"submitter": submitter,
	})

	downloadMsg := mq.NewDownloadMessage(site, id, submitter)
	msg := mq.NewMessage(ModuleName, mq.MsgDownload, downloadMsg)
	bot.sendChan <- *msg
}

func (bot *IrcBot) handleNowPlaying() {
	bot.log.Debug("Requesting NowPlaying")
	getNPMsg := mq.NewGetNowPlayingMessage()
	msg := mq.NewMessage(ModuleName, mq.MsgGetNowPlaying, getNPMsg)
	bot.sendChan <- *msg
	bot.nowPlayingRequested = true
}

func (bot *IrcBot) HandleIncreaseRating() {
	fname := "/dev/null"

	bot.log.WithFields(logrus.Fields{
		"fname": fname,
	}).Debug("Increasing rating")

	incRatingMsg := mq.NewIncreaseRatingMessage(fname)
	msg := mq.NewMessage(ModuleName, mq.MsgIncreaseRating, incRatingMsg)
	bot.sendChan <- *msg
}

func (bot *IrcBot) HandleDecreaseRating() {
	fname := "/dev/null"

	bot.log.WithFields(logrus.Fields{
		"fname": fname,
	})

	decRatingMsg := mq.NewDecreaseRatingMessage(fname)
	msg := mq.NewMessage(ModuleName, mq.MsgDecreaseRating, decRatingMsg)
	bot.sendChan <- *msg
}

func (bot *IrcBot) handleSongTooLong(id, submitter string) {
	response := fmt.Sprintf("%s: Your song with ID %s is too long for the stream", submitter, id)
	bot.conn.Privmsg(bot.cfg.IrcBot.Channel, response)
}

func (bot *IrcBot) handleDownloadAccepted(id, submitter string) {
	response := fmt.Sprintf("%s: Your song has been accepted for download, please hold", submitter)
	bot.conn.Privmsg(bot.cfg.IrcBot.Channel, response)
}

func (bot *IrcBot) handleDownloadCompleted(fname, submitter string) {
	response := fmt.Sprintf("%s: Added %s", submitter, fname)
	bot.conn.Privmsg(bot.cfg.IrcBot.Channel, response)
}

func (bot *IrcBot) handleSearchFailed(error, submitter string) {
	response := fmt.Sprintf("%s: Failed to search: %s", submitter, error)
	bot.conn.Privmsg(bot.cfg.IrcBot.Channel, response)
}

func (bot *IrcBot) handleQueueError(error, submitter string) {
	response := fmt.Sprintf("%s: Failed to search: %s", submitter, error)
	bot.conn.Privmsg(bot.cfg.IrcBot.Channel, response)
}

func (bot *IrcBot) handleQueueResponse(fname, submitter string, prio int) {
	response := fmt.Sprintf("%s: Queued %s at position %d", submitter, fname[:len(fname)-16], (mpd.MaxQueueSize - (prio - 2)))
	bot.conn.Privmsg(bot.cfg.IrcBot.Channel, response)
}
