package main

import (
	"flag"
	"github.com/r3boot/go-musicbot/pkg/artwork"
	"github.com/r3boot/go-musicbot/pkg/config"
	"github.com/r3boot/go-musicbot/pkg/id3tags"
	"github.com/sirupsen/logrus"
	"strings"

	"github.com/r3boot/go-musicbot/pkg/downloader"
	"github.com/r3boot/go-musicbot/pkg/indexer"
	"github.com/r3boot/go-musicbot/pkg/ircbot"
	"github.com/r3boot/go-musicbot/pkg/mpd"
	"github.com/r3boot/go-musicbot/pkg/mq"
	"github.com/r3boot/go-musicbot/pkg/rating"
)

const (
	defCfgFile  = "musicbot.yml"
	defLogJson  = false
	defLogLevel = "INFO"
)

var (
	msgQueue  *mq.MessageQueue
	mpdClient *mpd.MpdClient
	search    *indexer.Search
	download  *downloader.Download
	ratings   *rating.Ratings
	albumart  *artwork.AlbumArt
	tags      *id3tags.ID3Tags
	ircBot    *ircbot.IrcBot
	cfg       *config.MusicBotConfig

	log *logrus.Entry

	cfgFile  = flag.String("config", defCfgFile, "Configuration file to use")
	logJson  = flag.Bool("logjson", defLogJson, "Log output in JSON format")
	logLevel = flag.String("loglevel", defLogLevel, "Log Level to use")
)

func init() {
	var err error

	// Parse command-line flags
	flag.Parse()

	// Configure logging
	if *logJson {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	switch strings.ToUpper(*logLevel) {
	case "INFO":
		{
			logrus.SetLevel(logrus.InfoLevel)
		}
	case "DEBUG":
		{
			logrus.SetLevel(logrus.DebugLevel)
		}
	}

	// Initialize logging
	log = logrus.WithFields(logrus.Fields{
		"caller": "main",
	})

	// Load configuration
	cfg, err = config.LoadConfig(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	// Setup modules
	msgQueue, err = mq.NewMessageQueue()
	if err != nil {
		log.Fatal(err)
	}

	tags, err = id3tags.NewID3Tags(cfg)
	if err != nil {
		log.Fatal(err)
	}

	search, err = indexer.NewSearch(cfg)
	if err != nil {
		log.Fatal(err)
	}

	mpdClient, err = mpd.NewMpdClient(cfg, tags, search)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Features.IrcBot {
		ircBot, err = ircbot.NewIrcBot(cfg)
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg.Features.Download {
		download, err = downloader.NewDownloader(cfg, tags)
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg.Features.Ratings {
		ratings, err = rating.NewRatings()
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg.Features.API {
		albumart, err = artwork.NewAlbumArt()
		if err != nil {
			log.Fatal(err)
		}
	}

	// Setup Receiver channels
	if cfg.Features.IrcBot {
		err = msgQueue.NewReceiver(ircbot.ModuleName, ircBot.GetRecvChan())
		if err != nil {
			log.Fatal(err)
		}
	}

	err = msgQueue.NewReceiver(mpd.ModuleName, mpdClient.GetRecvChan())
	if err != nil {
		log.Fatal(err)
	}

	err = msgQueue.NewReceiver(indexer.ModuleName, search.GetRecvChan())
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Features.Download {
		err = msgQueue.NewReceiver(downloader.ModuleName, download.GetRecvChan())
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg.Features.Ratings {
		err = msgQueue.NewReceiver(rating.ModuleName, ratings.GetRecvChan())
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg.Features.API {
		err = msgQueue.NewReceiver(artwork.ModuleName, albumart.GetRecvChan())
		if err != nil {
			log.Fatal(err)
		}
	}

	// Setup Sender channels
	if cfg.Features.IrcBot {
		err = msgQueue.NewSender(ircbot.ModuleName, ircBot.GetSendChan())
		if err != nil {
			log.Fatal(err)
		}
	}

	err = msgQueue.NewSender(mpd.ModuleName, mpdClient.GetSendChan())
	if err != nil {
		log.Fatal(err)
	}

	err = msgQueue.NewSender(indexer.ModuleName, search.GetSendChan())
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Features.Download {
		err = msgQueue.NewSender(downloader.ModuleName, download.GetSendChan())
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg.Features.Ratings {
		err = msgQueue.NewSender(rating.ModuleName, ratings.GetSendChan())
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg.Features.API {
		err = msgQueue.NewSender(artwork.ModuleName, albumart.GetSendChan())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	if cfg.Features.IrcBot {
		ircBot.Run()
	}
}
