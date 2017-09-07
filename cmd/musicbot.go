package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/thoj/go-ircevent"
	"os/exec"
	"regexp"
)

const (
	D_SERVER   string = "irc.mononoke.nl"
	D_PORT     int    = 6697
	D_TLS      bool   = true
	D_NICKNAME string = "IrcBot"
	D_CHANNEL  string = "#zzz-someircbot"

	CMD_DJPLUS string = "!dj"
	CMD_NEXT   string = "!next"
)

const FETCH_YOUTUBE = "/usr/local/bin/fetch_youtube.sh"

var irccon *irc.Connection

var RE_CMD = regexp.MustCompile("^(\\![a-z]{2,4})")
var RE_DJHANDLER = regexp.MustCompile("(\\!dj\\+) ([a-zA-Z0-9_-]{11})")

var (
	server   = flag.String("server", D_SERVER, "Connect to this server")
	port     = flag.Int("port", D_PORT, "Port to connect to")
	useTLS   = flag.Bool("tls", D_TLS, "Enable TLS")
	nickname = flag.String("nick", D_NICKNAME, "Nickname to use")
	channel  = flag.String("chan", D_CHANNEL, "Channel to chill in")
)

func DownloadYID(yid string) {
	cmd := exec.Command(FETCH_YOUTUBE, yid)
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to run %s: %v\n", FETCH_YOUTUBE, err)
	}
}

func HandleYidDownload(channel, line string) {
	result := RE_DJHANDLER.FindAllStringSubmatch(line, -1)

	if len(result) == 1 {
		yid := result[0][2]
		go DownloadYID(yid)
		response := fmt.Sprintf("%s added to download queue", yid)
		irccon.Privmsg(channel, response)
	} else {
		fmt.Printf("no results found\n")
	}
}

func HandleNext(channel, line string) {
	fmt.Printf("Skipping to next song in playlist\n")
}

func ParsePrivmsg(e *irc.Event) {
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
	fmt.Printf("cmd: %s\n", cmd)

	switch cmd {
	case CMD_DJPLUS:
		HandleYidDownload(channel, line)
	case CMD_NEXT:
		HandleNext(channel, line)
	}
}

func main() {
	flag.Parse()

	fmt.Printf("Server: %s\n", *server)
	fmt.Printf("Port: %d\n", *port)
	fmt.Printf("TLS: %v\n", *useTLS)
	fmt.Printf("Nickname: %s\n", *nickname)
	fmt.Printf("Channel: %s\n", *channel)

	server := fmt.Sprintf("%s:%d", *server, *port)

	irccon = irc.IRC(*nickname, *nickname)
	irccon.VerboseCallbackHandler = true
	irccon.Debug = true
	irccon.UseTLS = *useTLS
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(*channel) })
	irccon.AddCallback("366", func(e *irc.Event) {})
	irccon.AddCallback("PRIVMSG", ParsePrivmsg)
	err := irccon.Connect(server)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	irccon.Loop()
}
