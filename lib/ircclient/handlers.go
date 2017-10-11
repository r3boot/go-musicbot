package ircclient

import (
	"fmt"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/thoj/go-ircevent"
)

func (c *IrcClient) initCallbacks() {
	c.conn.AddCallback("001", func(e *irc.Event) { c.conn.Join(c.config.IRC.Channel) })
	c.conn.AddCallback("366", func(e *irc.Event) {})
	c.conn.AddCallback("PRIVMSG", c.ParsePrivmsg)
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
	case CMD_DJPLUS:
		c.HandleYidDownload(channel, line)
	case CMD_PLAYLIST:
		c.HandlePlaylistDownload(channel, line)
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
	}
}

func (c *IrcClient) HandleYidDownload(channel, line string) {
	result := RE_DJHANDLER.FindAllStringSubmatch(line, -1)

	if len(result) == 1 {
		yid := result[0][2]
		c.ytClient.DownloadChan <- yid
		fmt.Printf("Added %s to download queue\n", yid)
	} else {
		fmt.Printf("no results found\n")
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

func (c *IrcClient) HandleNext(channel, line string) {
	fileName := c.mpdClient.Next()
	response := fmt.Sprintf("Now playing: %s", fileName)
	c.conn.Privmsg(channel, response)
}

func (c *IrcClient) HandleNowPlaying(channel, line string) {
	fileName := c.mpdClient.NowPlaying()
	if fileName == "" {
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
		response := fmt.Sprintf("Rating for %s is so low, it has been removed from the playlist", fileName)
		c.conn.Privmsg(channel, response)
	} else {
		response := fmt.Sprintf("Rating for %s is %d/10 .. BOOO!!!!", fileName, newRating)
		c.conn.Privmsg(channel, response)
	}
}

func (c *IrcClient) HandleIncreaseRating(channel, line string) {
	fileName := c.mpdClient.NowPlaying()
	newRating := c.mp3Library.IncreaseRating(fileName)
	fmt.Printf("IrcClient.HandleIncreaseRating rating for %s is now %d\n", fileName, newRating)
	response := fmt.Sprintf("Rating for %s is %d/10 .. Party on!!!!", fileName, newRating)
	c.conn.Privmsg(channel, response)
}
