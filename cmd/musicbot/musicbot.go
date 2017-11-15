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
	"github.com/r3boot/go-musicbot/lib/webapi"
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

	MP3Library := mp3lib.NewMP3Library(musicDir)

	MPDClient, err := mpdclient.NewMPDClient(Config, MP3Library)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MPDClient: %v\n", err)
		os.Exit(1)
	}

	YoutubeClient := youtubeclient.NewYoutubeClient(Config, MPDClient, MP3Library, musicDir)

	if Config.App.Daemonize {
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

	// API + Web UI
	if Config.App.APIEnabled {
		WebApi := webapi.NewWebApi(Config, MPDClient, MP3Library, YoutubeClient)

		if err = WebApi.Setup(); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}

		if Config.App.IrcBotEnabled {
			go WebApi.Run()
		} else {
			WebApi.Run()
		}
	}

	// IRC bot
	if Config.App.IrcBotEnabled {
		IRCClient := ircclient.NewIrcClient(Config, MPDClient, YoutubeClient, MP3Library)
		if err = IRCClient.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to run IRC bot: %v", err)
			os.Exit(1)
		}
	}
}
