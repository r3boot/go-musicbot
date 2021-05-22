package config

import (
	"fmt"
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

type WebUi struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
	Token   string `yaml:"token"`
}

type Config struct {
	Datastore  *DatastoreConfig `yaml:"datastore"`
	Mpd        *MpdConfig       `yaml:"mpd"`
	Liquidsoap *LSConfig        `yaml:"liquidsoap"`
	Postgres   *PostgresConfig  `yaml:"postgres"`
	Youtube    *YoutubeConfig   `yaml:"youtube"`
	WebApi     *WebApi          `yaml:"webapi"`
	WebUi      *WebUi           `yaml:"webui"`
}

func New(fname string) (*Config, error) {
	cfg := &Config{}

	err := cfg.Load(fname)
	if err != nil {
		return nil, fmt.Errorf("cfg.Load: %v", err)
	}

	return cfg, nil
}

func (c *Config) Load(fname string) error {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile: %v\n", err)
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return fmt.Errorf("yaml.Unmarshal: %v\n", err)
	}

	return nil
}
