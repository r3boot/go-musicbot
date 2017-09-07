package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/thoj/go-ircevent"
	"gopkg.in/sevlyar/go-daemon.v0"
)

const (
	D_SERVER    string = "irc.mononoke.nl"
	D_PORT      int    = 6697
	D_TLS       bool   = true
	D_NICKNAME  string = "IrcBot"
	D_CHANNEL   string = "#zzz-someircbot"
	D_DAEMONIZE bool   = false
	D_BASEDIR   string = "/music"
	D_YTDL      string = "/usr/local/bin/youtube-dl"

	YOUTUBE_URL  string = "https://www.youtube.com/watch?v="
	YOUTUBE_SEEN string = "/var/spool/fetch_youtube.seen"

	CMD_DJPLUS string = "!dj"
	CMD_NEXT   string = "!next"
)

var irccon *irc.Connection

var RE_CMD = regexp.MustCompile("^(\\![a-z]{2,4})")
var RE_DJHANDLER = regexp.MustCompile("(\\!dj\\+) ([a-zA-Z0-9_-]{11})")

var (
	server    = flag.String("server", D_SERVER, "Connect to this server")
	port      = flag.Int("port", D_PORT, "Port to connect to")
	useTLS    = flag.Bool("tls", D_TLS, "Enable TLS")
	nickname  = flag.String("nick", D_NICKNAME, "Nickname to use")
	channel   = flag.String("chan", D_CHANNEL, "Channel to chill in")
	daemonize = flag.Bool("d", D_DAEMONIZE, "Daemonize process")
	baseDir   = flag.String("dir", D_BASEDIR, "Basedir under which to write files")
	ytdl      = flag.String("ytdl", D_YTDL, "Path to youtube-dl")

	musicDir string
)

func hasYID(yid string) bool {
	fd, err := os.Open(YOUTUBE_SEEN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		return false
	}

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		if scanner.Text() == yid {
			return true
		}
	}

	return false
}

func stripChannel(channel string) string {
	result := ""
	for i := 0; i < len(channel); i++ {
		if channel[i] == '#' {
			continue
		}
		result += string(channel[i])
	}

	return result
}

func DownloadYID(yid string) {
	if hasYID(yid) {
		fmt.Printf("YID %s has already been downloaded\n", yid)
	}
	output := fmt.Sprintf("%s/%%(title)s-%%(id)s.%%(ext)s", musicDir)
	url := fmt.Sprintf("%s%s", YOUTUBE_URL, yid)
	cmd := exec.Command(*ytdl, "-x", "--audio-format", "mp3", "-o", output, url)
	fmt.Printf("Running command: %v\n", cmd)
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to run %s: %v\n", *ytdl, err)
	}
	cmd.Wait()
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

func RunIrcBot() {
	server := fmt.Sprintf("%s:%d", *server, *port)

	irccon = irc.IRC(*nickname, *nickname)
	irccon.VerboseCallbackHandler = true
	irccon.Debug = false
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

func main() {
	flag.Parse()

	chanName := stripChannel(*channel)
	musicDir = fmt.Sprintf("%s/%s", *baseDir, chanName)

	pidFile := fmt.Sprintf("/var/musicbot/%s-%s.pid", *nickname, chanName)
	logFile := fmt.Sprintf("/var/log/musicbot/%s-%s.log", *nickname, chanName)

	if *daemonize {
		ctx := daemon.Context{
			PidFileName: pidFile,
			PidFilePerm: 0644,
			LogFileName: logFile,
			LogFilePerm: 0640,
			WorkDir:     "/tmp",
			Umask:       022,
			Args:        []string{},
		}

		d, err := ctx.Reborn()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to run as daemon: %v", err)
			os.Exit(1)
		}
		if d != nil {
			return
		}
		defer ctx.Release()
	}

	RunIrcBot()
}
