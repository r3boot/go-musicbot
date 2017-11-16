package config

import (
	"fmt"
	"io/ioutil"

	"github.com/r3boot/go-musicbot/lib/logger"
	yaml "gopkg.in/yaml.v2"
)

var log *logger.Logger

func LoadConfig(l *logger.Logger, filename string) (config *MusicBotConfig, err error) {
	log = l
	config = &MusicBotConfig{}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig failed: %v", err)
	}

	if err = yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("config.LoadConfig failed: %v", err)
	}

	return config, nil
}
