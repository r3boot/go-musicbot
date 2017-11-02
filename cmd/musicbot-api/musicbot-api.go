package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"github.com/r3boot/go-musicbot/lib/webapi"
	"github.com/r3boot/go-musicbot/lib/ytclient"
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

	WebApi := webapi.NewWebApi(Config, MPDClient, MP3Library, YoutubeClient)

	if err = WebApi.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	WebApi.Run()
}
