package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/r3boot/test/lib/webapi"

	"github.com/sirupsen/logrus"

	"github.com/r3boot/test/lib/config"
)

const (
	defConfigFileValue = "/etc/musicbot.yml"
	defHostValue       = "0.0.0.0"
	defPortValue       = 8080
	helpConfigFile     = "Path to the configuration file"
	helpHost           = "Host to listen on"
	helpPort           = "Port to listen on"
)

func Error(err error) {
	fmt.Printf("ERROR: %v\n", err)
	os.Exit(1)
}

func main() {
	var (
		ConfigFile = flag.String("config", defConfigFileValue, helpConfigFile)
		LogLevel   = flag.String("loglevel", "INFO", "Log level to use (INFO, DEBUG)")
		LogJson    = flag.Bool("json", false, "Output logging in JSON format")
		Host       = flag.String("host", defHostValue, helpHost)
		Port       = flag.Int("port", defPortValue, helpPort)

		log *logrus.Entry
	)
	flag.Parse()

	// Configure logging
	if *LogJson {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	switch strings.ToUpper(*LogLevel) {
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

	cfg, err := config.New(*ConfigFile)
	if err != nil {
		log.Fatalf("config.New: %v", err)
	}

	api, err := webapi.NewWebApi(cfg)
	if err != nil {
		log.Fatalf("NewWebApi: %v", err)
	}

	log.Debugf("Host: %s; Port: %d; api: %v", Host, Port, api)

	err = api.Run(*Host, *Port)
	if err != nil {
		log.Fatalf("RunWebApi: %v", err)
	}
}
