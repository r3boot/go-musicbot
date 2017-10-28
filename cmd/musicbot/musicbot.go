package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mpdclient"

	"github.com/r3boot/go-musicbot/lib/ircclient"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/ytclient"
	"gopkg.in/sevlyar/go-daemon.v0"
)

const (
	D_CFGFILE string = "musicbot.yaml"
)

var (
	cfgFile  = flag.String("f", D_CFGFILE, "Configuration file to use")
	musicDir string
)



func main() {
	var err error

	flag.Parse()

	rand.Seed(time.Now().Unix())

	Config, err := config.LoadConfig(*cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	chanName := stripChannel(Config.IRC.Channel)
	musicDir = fmt.Sprintf("%s/%s", Config.Youtube.BaseDir, chanName)

	MPDClient, err := mpdclient.NewMPDClient(Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MPDClient: %v\n", err)
		os.Exit(1)
	}

	MP3Library := mp3lib.NewMP3Library(musicDir)

	YoutubeClient := youtubeclient.NewYoutubeClient(Config, MPDClient, MP3Library, musicDir)

	IRCClient := ircclient.NewIrcClient(Config, MPDClient, YoutubeClient, MP3Library)

	if Config.IRC.Daemonize {
		pidFile := fmt.Sprintf("/var/musicbot/%s-%s.pid", Config.IRC.Nickname, chanName)
		logFile := fmt.Sprintf("/var/log/musicbot/%s-%s.log", Config.IRC.Nickname, chanName)

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

	if err = IRCClient.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run IRC bot: %v", err)
		os.Exit(1)
	}
}
