package config

import (
	"fmt"
	"github.com/r3boot/go-musicbot/lib/log"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	AllowPlayerNext uint64 = 1 << iota
	AllowPlayerNowPlaying
	AllowPlayerQueue
	AllowRatingIncrease
	AllowRatingDecrease
	AllowTrackSearch
	AllowTrackRequest
	AllowTrackDownload
	AllowSynchronize
)

type DatastoreConfig struct {
	Directory string `yaml:"directory"`
}

type PostgresConfig struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type MpdConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
}

type LSConfig struct {
	Address    string `yaml:"address"`
	OutputName string `yaml:"output"`
	QueueName  string `yaml:"queue"`
}

type YoutubeConfig struct {
	BaseUrl          string `yaml:"base_url"`
	Binary           string `yaml:"binary"`
	TmpDir           string `yaml:"tmpdir"`
	MaxAllowedLength int    `yaml:"max_allowed_length"`
}

type ApiUser struct {
	Name           string   `yaml:"name"`
	Token          string   `yaml:"token"`
	Authorizations []string `yaml:"authorizations"`
}

type WebApi struct {
	Address string    `yaml:"address"`
	Port    int       `yaml:"port"`
	Users   []ApiUser `yaml:"users"`
}

type Discogs struct {
	CacheDir string `yaml:"cache_dir"`
	Token    string `yaml:"token"`
}

type WebUi struct {
	Address string   `yaml:"address"`
	Port    int      `yaml:"port"`
	Token   string   `yaml:"token"`
	Discogs *Discogs `yaml:"discogs"`
}

type IrcBot struct {
	Nickname         string   `yaml:"nickname"`
	Server           string   `yaml:"server"`
	Port             int      `yaml:"port"`
	Channel          string   `yaml:"channel"`
	Tls              bool     `yaml:"tls"`
	VerifyTls        bool     `yaml:"verify_tls"`
	Token            string   `yaml:"token"`
	CommandCharacter string   `yaml:"command_character"`
	ValidCommands    []string `yaml:"valid_commands"`
	StreamUrl        string   `yaml:"stream_url"`
	RadioReplies     []string `yaml:"radio_replies"`
	Ch00nReplies     []string `yaml:"ch00n_replies"`
}

type Config struct {
	Datastore  *DatastoreConfig `yaml:"datastore"`
	Mpd        *MpdConfig       `yaml:"mpd"`
	Liquidsoap *LSConfig        `yaml:"liquidsoap"`
	Postgres   *PostgresConfig  `yaml:"postgres"`
	Youtube    *YoutubeConfig   `yaml:"youtube"`
	WebApi     *WebApi          `yaml:"webapi"`
	WebUi      *WebUi           `yaml:"webui"`
	IrcBot     *IrcBot          `yaml:"ircbot"`
}

func New(fname string) (*Config, error) {
	cfg := &Config{}

	err := cfg.Load(fname)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Load(fname string) error {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "config",
			"function": "Load",
			"call":     "ioutils.ReadFile",
			"filename": fname,
		}, err.Error())
		return fmt.Errorf("failed to load configfile")
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "config",
			"function": "Load",
			"call":     "yaml.Unmarshal",
			"filename": fname,
		}, err.Error())
		return fmt.Errorf("failed to parse configfile")
	}

	return nil
}
