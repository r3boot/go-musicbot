package config

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

func LoadConfig(filename string) (config *MusicBotConfig, err error) {
	config = &MusicBotConfig{}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig failed: %v", err)
	}

	if err = yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("LoadConfig failed: %v", err)
	}

	return config, nil
}