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
	"github.com/r3boot/go-musicbot/lib/logger"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/webapi"
	"github.com/r3boot/go-musicbot/lib/ytclient"
	"gopkg.in/sevlyar/go-daemon.v0"
)

const (
	D_CFGFILE string = "musicbot.yaml"
	D_DEBUG   bool   = false
)

var (
	cfgFile  = flag.String("f", D_CFGFILE, "Configuration file to use")
	debug    = flag.Bool("d", D_DEBUG, "Enable debug mode")
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

	Logger := logger.NewLogger(*debug, *debug)

	Config, err := config.LoadConfig(Logger, *cfgFile)
	if err != nil {
		Logger.Fatalf("%v", err)
	}

	chanName := stripChannel(Config.IRC.Channel)
	musicDir = fmt.Sprintf("%s/%s", Config.Youtube.BaseDir, chanName)

	MP3Library := mp3lib.NewMP3Library(Logger, musicDir)

	MPDClient, err := mpdclient.NewMPDClient(Logger, Config, MP3Library)
	if err != nil {
		Logger.Fatalf("%v", err)
	}

	YoutubeClient := youtubeclient.NewYoutubeClient(Logger, Config, MPDClient, MP3Library, musicDir)

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
			Logger.Fatalf("Unable to run as daemon: %v", err)
		}
		if d != nil {
			return
		}
		defer ctx.Release()
	}

	// API + Web UI
	if Config.App.APIEnabled {
		WebApi := webapi.NewWebApi(Logger, Config, MPDClient, MP3Library, YoutubeClient)

		if err = WebApi.Setup(); err != nil {
			Logger.Fatalf("%v", err)
		}

		if Config.App.IrcBotEnabled {
			Logger.Debugf("WebUI enabled")
			go WebApi.Run()
		} else {
			WebApi.Run()
		}
	}

	// IRC bot
	if Config.App.IrcBotEnabled {
		IRCClient := ircclient.NewIrcClient(Logger, Config, MPDClient, YoutubeClient, MP3Library)
		Logger.Debugf("Running IRC bot")
		if err = IRCClient.Run(); err != nil {
			Logger.Fatalf("%v", err)
			os.Exit(1)
		}
	}
}
