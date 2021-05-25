package main

import (
	"flag"
	"strings"

	"github.com/r3boot/go-musicbot/lib/webapi"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/log"
)

const (
	defConfigFileValue = "/etc/musicbot.yml"
	defHostValue       = "0.0.0.0"
	defPortValue       = 8080
	defLogLevel        = "info"
	defJson            = false

	helpConfigFile = "Path to the configuration file"
	helpHost       = "Host to listen on"
	helpPort       = "Port to listen on"
	helpLogLevel   = "Log level to use (info, debug)"
	helpJson       = "Output logging in JSON format"
)

func main() {
	var (
		ConfigFile = flag.String("config", defConfigFileValue, helpConfigFile)
		LogLevel   = flag.String("loglevel", defLogLevel, helpLogLevel)
		LogJson    = flag.Bool("json", defJson, helpJson)
		Host       = flag.String("host", defHostValue, helpHost)
		Port       = flag.Int("port", defPortValue, helpPort)
	)
	flag.Parse()

	log.NewLogger(strings.ToUpper(*LogLevel), *LogJson)

	cfg, err := config.New(*ConfigFile)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "main",
			"function": "main",
			"call":     "config.New",
			"filename": *ConfigFile,
		}, err.Error())
	}

	api, err := webapi.NewWebApi(cfg)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "main",
			"function": "main",
			"call":     "webapi.NewWebApi",
		}, err.Error())
	}

	log.Debugf(log.Fields{
		"package":  "main",
		"function": "main",
		"host":     *Host,
		"port":     *Port,
	}, "starting api")

	err = api.Run(*Host, *Port)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "main",
			"function": "main",
			"call":     "api.Run",
			"host":     *Host,
			"port":     *Port,
		}, err.Error())
	}
}
