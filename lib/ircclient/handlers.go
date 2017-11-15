package ircclient

import (
	"fmt"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/thoj/go-ircevent"
	"os"
	"strings"
	"time"
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

	cmd := cmdResult[0][1]

	command, ok := c.isValidCommand(cmd)
	if !ok {
		return
	}

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
	case CMD_PLAY:
		c.HandleSearchAndPlay(channel, line)
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
		fmt.Printf("Added %s to download queue\n", yid)
		if !c.Online[NICK_SJAAK] {
			response = fmt.Sprintf("Added %s%s to the download queue", c.config.Youtube.BaseUrl, yid)
			c.conn.Privmsg(channel, response)
			privmsgResponse := fmt.Sprintf("Jow! Sjaak is offline! Denk nou aan je SLA JONGUH! Voeg dit ff toe aan de DB: %s", yid)
			c.conn.Privmsg(NICK_FLUNK, privmsgResponse)
		}
	} else {
		response = fmt.Sprintf("No yid found in message .. Anta BAKA??")
		fmt.Printf("no results found\n")
		c.conn.Privmsg(channel, response)
	}
}

func (c *IrcClient) HandlePlaylistDownload(channel, line string) {
	result := RE_DJLIST.FindAllStringSubmatch(line, -1)

	if len(result) == 1 {
		playlistUrl := result[0][2]
		c.ytClient.PlaylistChan <- playlistUrl
		fmt.Printf("Added playlist %s to download queue\n", playlistUrl)
		response := fmt.Sprintf("Added playlist to download queue")
		c.conn.Privmsg(channel, response)
	} else {
		fmt.Printf("no playlist found\n")
		response := fmt.Sprintf("Did not find any playlist")
		c.conn.Privmsg(channel, response)
	}
}

func (c *IrcClient) HandleStart(channel, line string) {
	c.mpdClient.Shuffle()
	fileName := c.mpdClient.Play()
	response := fmt.Sprintf("Now playing: %s", fileName)
	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleNext(channel, line string) {
	c.HandleDecreaseRating(channel, line)
	fileName := c.mpdClient.Next()
	response := fmt.Sprintf("Now playing: %s", fileName)
	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleNowPlaying(channel, line string) {
	fileName := c.mpdClient.NowPlaying()
	if strings.HasPrefix(fileName, "Error: ") {
		fileName = c.mpdClient.Play()
	}

	duration := c.mpdClient.Duration()
	rating := c.mp3Library.GetRating(fileName)

	response := fmt.Sprintf("Now playing: %s (duration: %s; rating: %d/10)", fileName[:len(fileName)-16], duration, rating)
	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleRadioUrl(channel, line string) {
	response := fmt.Sprintf("%s Listen to %s", c.randomRadioMessage(), c.config.Bot.StreamURL)
	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleDecreaseRating(channel, line string) {
	fileName := c.mpdClient.NowPlaying()
	newRating := c.mp3Library.DecreaseRating(fileName)
	fmt.Printf("IrcClient.HandleDecreaseRating rating for %s is now %d\n", fileName, newRating)
	if newRating == mp3lib.RATING_ZERO {
		c.mpdClient.Next()
		c.mp3Library.RemoveFile(fileName)
		response := fmt.Sprintf("Rating for %s is so low, it has been removed from the playlist", fileName[:len(fileName)-16])
		c.conn.Privmsg(channel, response)
	} else {
		response := fmt.Sprintf("Rating for %s is %d/10 .. BOOO!!!!", fileName[:len(fileName)-16], newRating)
		c.conn.Privmsg(channel, response)
	}
}

func (c *IrcClient) HandleIncreaseRating(channel, line string) {
	fileName := c.mpdClient.NowPlaying()
	newRating := c.mp3Library.IncreaseRating(fileName)
	fmt.Printf("IrcClient.HandleIncreaseRating rating for %s is now %d\n", fileName, newRating)
	response := fmt.Sprintf("Rating for %s is %d/10 .. Party on!!!!", fileName[:len(fileName)-16], newRating)
	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleSearchAndPlay(channel, line string) {
	result := RE_SEARCH.FindAllStringSubmatch(line, -1)
	response := "Undefined"

	if len(result) == 1 {
		if len(result[0][2]) > 256 {
			response = fmt.Sprintf("Size of query too large")
			c.conn.Privmsg(channel, response)
		}

		query := result[0][2]

		qpos, err := c.mpdClient.Enqueue(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to enqueue: %v\n", err)
			response = fmt.Sprintf("Failed to enqueue: %v", err)
		} else {
			title, err := c.mpdClient.GetTitle(qpos)
			if err != nil {
				response = fmt.Sprintf("Added to queue")
			} else {
				response = fmt.Sprintf("Added %s to queue", title)
			}
		}
	} else {
		response = fmt.Sprintf("Need a query to search .. stupid!")
	}
	c.conn.Privmsg(channel, response)
}
