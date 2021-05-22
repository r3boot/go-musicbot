package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	ModuleName = "Config"
)

type Features struct {
	PartyMode bool `yaml:"partymode"`
	Queue     bool `yaml:"queue"`
	Download  bool `yaml:"download"`
	Ratings   bool `yaml:"ratings"`
	IrcBot    bool `yaml:"ircbot"`
	API       bool `yaml:"api"`
	WebUI     bool `yaml:"webui"`
	Daemonize bool `yaml:"daemonize"`
	Debug     bool `yaml:"debug"`
}

type Paths struct {
	Music     string `yaml:"music"`
	TmpDir    string `yaml:"tmpdir"`
	Index     string `yaml:"index"`
	Youtubedl string `yaml:"youtubedl"`
	Id3v2     string `yaml:"id3v2"`
}

type IrcBot struct {
	Nickname      string   `yaml:"nickname"`
	Server        string   `yaml:"server"`
	Port          int      `yaml:"port"`
	Channel       string   `yaml:"channel"`
	UseTLS        bool     `yaml:"tls"`
	VerifyTLS     bool     `yaml:"tls_verify"`
	Debug         bool     `yaml:"debug"`
	CommandChar   string   `yaml:"command_character"`
	ValidCommands []string `yaml:"valid_commands"`
	StreamURL     string   `yaml:"stream_url"`
	RadioMsgs     []string `yaml:"radio_messages"`
	Ch00nMsgs     []string `yaml:"ch00n_messages"`
}

type Download struct {
	NumWorkers int    `yaml:"num_workers"`
	Url        string `yaml:"url"`
}

type MPD struct {
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type Api struct {
	Address      string `yaml:"address"`
	Port         string `yaml:"port"`
	Title        string `yaml:"title"`
	OggStreamURL string `yaml:"ogg_stream_url"`
	Mp3StreamURL string `yaml:"mp3_stream_url"`
	Assets       string `yaml:"assets"`
}

type MusicBotConfig struct {
	Features Features `yaml:"features"`
	Paths    Paths    `yaml:"paths"`
	IrcBot   IrcBot   `yaml:"ircbot"`
	Download Download `yaml:"download"`
	MPD      MPD      `yaml:"mpd"`
	Api      Api      `yaml:"api"`
}

var (
	log *logrus.Entry
)

func LoadConfig(fname string) (*MusicBotConfig, error) {
	log = logrus.WithFields(logrus.Fields{
		"caller": ModuleName,
		"fname":  fname,
	})

	config := &MusicBotConfig{}

	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadFile: %v", err)
	}

	if err = yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %v", err)
	}

	log.Debug("Configuration loaded")

	log.WithFields(logrus.Fields{
		"enabled": config.Features.PartyMode,
	}).Debug("Feature PartyMode")

	log.WithFields(logrus.Fields{
		"enabled": config.Features.Queue,
	}).Debug("Feature Queue")

	log.WithFields(logrus.Fields{
		"enabled": config.Features.Download,
	}).Debug("Feature Download")

	log.WithFields(logrus.Fields{
		"enabled": config.Features.Ratings,
	}).Debug("Feature Ratings")

	log.WithFields(logrus.Fields{
		"enabled": config.Features.IrcBot,
	}).Debug("Feature IrcBot")

	log.WithFields(logrus.Fields{
		"enabled": config.Features.API,
	}).Debug("Feature API")

	log.WithFields(logrus.Fields{
		"enabled": config.Features.WebUI,
	}).Debug("Feature WebUI")

	return config, nil
}
