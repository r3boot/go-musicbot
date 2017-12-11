package ircclient

import (
	"fmt"
	"strings"
	"time"

	"go-ircevent"

	"github.com/r3boot/go-musicbot/lib/id3tags"
)

func (c *IrcClient) initCallbacks() {
	c.conn.AddCallback("001", func(e *irc.Event) { c.conn.Join(c.config.IRC.Channel) })
	c.conn.AddCallback("366", func(e *irc.Event) {}) // RPL_ENDOFNAMES
	c.conn.AddCallback("303", c.ParseIsOn)           // RPL_ISON
	c.conn.AddCallback("PRIVMSG", c.ParsePrivmsg)
}

func (c *IrcClient) CheckIfSjaakIsOnline() {
	c.Online[NICK_SJAAK] = false
	for {
		c.conn.SendRawf("ISON %s", NICK_SJAAK)
		time.Sleep(5 * time.Second)
	}
}

func (c *IrcClient) ParseIsOn(e *irc.Event) {
	if len(e.Arguments) != 2 {
		log.Warningf("IrcClient.ParseIsOn: Invalid number of arguments")
		return
	}

	line := e.Arguments[1]

	if len(line) > 0 {
		isOnNickname := strings.Split(line, " ")[0]
		if isOnNickname == NICK_SJAAK {
			c.Online[NICK_SJAAK] = true
		}
	} else {
		c.Online[NICK_SJAAK] = false
	}
}

func (c *IrcClient) ParsePrivmsg(e *irc.Event) {
	if len(e.Arguments) != 2 {
		return
	}

	channel := e.Arguments[0]
	line := e.Arguments[1]

	cmdResult := RE_CMD.FindAllStringSubmatch(line, -1)
	if len(cmdResult) != 1 {
		return
	}

	command := cmdResult[0][1]

	log.Debugf("IrcClient.ParsePrivmsg: Got command %s", command)

	switch command {
	case CMD_HELP:
		c.HandleHelp(channel, line)
	case CMD_DJPLUS:
		c.HandleYidDownload(channel, line)
	case CMD_PLAYLIST:
		c.HandlePlaylistDownload(channel, line)
	case CMD_START:
		c.HandleStart(channel, line)
	case CMD_NEXT:
		c.HandleNext(channel, line)
	case CMD_PLAYING:
		c.HandleNowPlaying(channel, line)
	case CMD_RADIO:
		c.HandleRadioUrl(channel, line)
	case CMD_BOO:
		c.HandleDecreaseRating(channel, line)
	case CMD_TUNE:
		c.HandleIncreaseRating(channel, line)
	case CMD_REQUEST:
		c.HandleSearchAndPlay(channel, line)
	case CMD_QUEUE:
		c.HandleShowQueue(channel, line)
	default:
		log.Warningf("IrcClient.ParsePrivmsg: Invalid command received: %s", command)
	}
}

func (c *IrcClient) HandleHelp(channel, line string) {
	c.conn.Privmsg(channel, RESPONSE_HELP)
}

func (c *IrcClient) HandleYidDownload(channel, line string) {
	result := RE_DJHANDLER.FindAllStringSubmatch(line, -1)

	response := "Undefined"

	if len(result) == 1 {
		yid := result[0][2]
		c.ytClient.DownloadChan <- yid
		log.Infof("Added %s to download queue", yid)
		if !c.Online[NICK_SJAAK] {
			response = fmt.Sprintf("Added %s%s to the download queue", c.config.Youtube.BaseUrl, yid)
			c.conn.Privmsg(channel, response)
			privmsgResponse := fmt.Sprintf("Jow! Sjaak is offline! Denk nou aan je SLA JONGUH! Voeg dit ff toe aan de DB: %s", yid)
			c.conn.Privmsg(NICK_FLUNK, privmsgResponse)
			log.Infof("Sent message to %s", NICK_FLUNK)
		}
	} else {
		response = fmt.Sprintf("No yid found in message .. Anta BAKA??")
		log.Warningf("IrcClient.HandleYidDownload: no results found")
		c.conn.Privmsg(channel, response)
	}
}

func (c *IrcClient) HandlePlaylistDownload(channel, line string) {
	result := RE_DJLIST.FindAllStringSubmatch(line, -1)

	if len(result) == 1 {
		playlistUrl := result[0][2]
		c.ytClient.PlaylistChan <- playlistUrl
		log.Infof("Added playlist %s to download queue", playlistUrl)
		response := fmt.Sprintf("Added playlist to download queue")
		c.conn.Privmsg(channel, response)
	} else {
		log.Warningf("IrcClient.HandlePlaylistDownload: no playlist found")
		response := fmt.Sprintf("Did not find any playlist")
		c.conn.Privmsg(channel, response)
	}
}

func (c *IrcClient) HandleStart(channel, line string) {
	c.mpdClient.Shuffle()
	np := c.mpdClient.Play()

	response := fmt.Sprintf("Now playing: %s", np.Title)
	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleNext(channel, line string) {
	c.HandleDecreaseRating(channel, line)
	np := c.mpdClient.Next()
	response := fmt.Sprintf("Now playing: %s", np.Title)
	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleNowPlaying(channel, line string) {
	np := c.mpdClient.NowPlaying()
	fname := np.Filename
	if strings.HasPrefix(fname, "Error: ") {
		fname = c.mpdClient.Play().Filename
	}

	duration := np.Duration
	rating, err := c.id3.GetRating(fname)
	if err != nil {
		log.Warningf("IrcClient.HandleNowPlaying: %v", err)
		response := "Failed to retrieve NowPlaying data"
		c.conn.Privmsg(channel, response)
		return
	}

	durationFormat := fmt.Sprintf("%ds", int(duration))
	d, _ := time.ParseDuration(durationFormat)
	formattedDuration := d.String()

	response := fmt.Sprintf("Now playing: %s (duration: %s; rating: %d/10)", fname[:len(fname)-16], formattedDuration, rating)
	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleRadioUrl(channel, line string) {
	response := fmt.Sprintf("%s Listen to %s", c.randomRadioMessage(), c.config.Bot.StreamURL)
	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleDecreaseRating(channel, line string) {
	np := c.mpdClient.NowPlaying()
	fname := np.Filename
	newRating, err := c.id3.DecreaseRating(fname)
	if err != nil {
		log.Warningf("IrcClient.HandleDecreaseRating: %v", err)
		response := "Failed to decrease rating"
		c.conn.Privmsg(channel, response)
		return
	}

	log.Infof("Rating for %s is now %d", fname, newRating)
	if newRating == id3tags.RATING_ZERO {
		c.mpdClient.Next()
		err := c.id3.RemoveFile(fname)
		if err != nil {
			log.Warningf("IrcClient.HandleDecreaseRating: %v", err)
			response := "Failed to decrease rating"
			c.conn.Privmsg(channel, response)
			return
		}
		log.Warningf("IrcClient.HandleDecreaseRating: Rating was 0, removed %s", fname)
		response := fmt.Sprintf("Rating for %s is so low, it has been removed from the playlist", fname[:len(fname)-16])
		c.conn.Privmsg(channel, response)
	} else {
		response := fmt.Sprintf("Rating for %s is %d/10 .. BOOO!!!!", fname[:len(fname)-16], newRating)
		c.conn.Privmsg(channel, response)
	}
}

func (c *IrcClient) HandleIncreaseRating(channel, line string) {
	np := c.mpdClient.NowPlaying()
	fname := np.Filename
	newRating, err := c.id3.IncreaseRating(fname)
	if err != nil {
		log.Warningf("IrcClient.HandleIncreaseRating: %v", err)
		response := "Failed to increase rating"
		c.conn.Privmsg(channel, response)
		return
	}

	log.Infof("Rating for %s is now %d", fname, newRating)
	response := fmt.Sprintf("Rating for %s is %d/10 .. Party on!!!!", fname[:len(fname)-16], newRating)
	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleSearchAndPlay(channel, line string) {
	result := RE_SEARCH.FindAllStringSubmatch(line, -1)
	response := "Undefined"

	if len(result) == 1 {
		if len(result[0][2]) > 256 {
			log.Warningf("IrcClient.HandleSearchAndPlay: query too large")
			response = fmt.Sprintf("Size of query too large")
			c.conn.Privmsg(channel, response)
			return
		}

		query := result[0][2]

		entry, err := c.mpdClient.Enqueue(query)
		qpos := entry.QPrio
		if err != nil {
			log.Warningf("IrcClient.HandleSearchAndPlay: %v", err)
			response = fmt.Sprintf("Failed to enqueue: %v", err)
		} else {
			title, err := c.mpdClient.GetTitle(qpos)
			if err != nil {
				log.Warningf("IrcClient.HandleSearchAndPlay: %v", err)
				response = fmt.Sprintf("Failed to get title")
			} else {
				name := title[:len(title)-16]
				log.Infof("Added %s to the queue", name)
				response = fmt.Sprintf("Added %s to the queue", name)
			}
		}
	} else {
		log.Warningf("IrcClient.HandleSearchAndPlay: No query found")
		response = fmt.Sprintf("Need a query to search .. stupid!")
	}

	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleShowQueue(channel, line string) {
	entries := c.mpdClient.GetPlayQueue()

	if len(entries) == 0 {
		response := fmt.Sprintf("Queue is empty")
		c.conn.Privmsg(channel, response)
		return
	}

	c.conn.Privmsg(channel, "Current queue:")

	for idx := 0; idx < len(entries); idx++ {
	}

	for _, idx := range entries.Keys() {
		entry := entries[idx]
		response := ""
		if entry.Artist != "" {
			response = fmt.Sprintf("%d) %s - %s\n", idx, entry.Artist, entry.Title)
		} else {
			response = fmt.Sprintf("%d) %s\n", idx, entry.Title)
		}
		c.conn.Privmsg(channel, response)
	}

	log.Debugf("%v", entries)
}
