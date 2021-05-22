package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/r3boot/test/lib/config"
	"github.com/r3boot/test/lib/webui"

	"github.com/go-openapi/runtime"

	"github.com/go-openapi/strfmt"

	httptransport "github.com/go-openapi/runtime/client"

	"github.com/r3boot/test/lib/apiclient"

	"github.com/sirupsen/logrus"
)

const (
	defHostValue       = "localhost"
	defPortValue       = 8080
	defConfigFileValue = "/etc/musicbot.yml"

	helpHost       = "Host to connect to"
	helpPort       = "Port to connect to"
	helpConfigFile = "Path to the configuration file"

	musicbotSubmitter = "musicbot"

	envVarPrefix = "MUSICBOT_WEBUI_"
	envVarHost   = envVarPrefix + "HOST"
	envVarPort   = envVarPrefix + "PORT"
	envVarToken  = envVarPrefix + "TOKEN"
)

func argOrEnvVar(argValue interface{}, envVarName string) (interface{}, error) {
	result := argValue
	envValue, ok := os.LookupEnv(envVarName)
	if ok {
		result = envValue
	}

	if result == "" {
		return "", fmt.Errorf("No value found")
	}

	return result, nil
}

func main() {
	var (
		Host       = flag.String("host", defHostValue, helpHost)
		Port       = flag.Int("port", defPortValue, helpPort)
		LogLevel   = flag.String("loglevel", "INFO", "Log level to use (INFO, DEBUG)")
		LogJson    = flag.Bool("json", false, "Output logging in JSON format")
		ConfigFile = flag.String("config", defConfigFileValue, helpConfigFile)

		Token = flag.String("token", "", "Authentication token to use")

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
		"module": "main",
	})

	host, err := argOrEnvVar(*Host, envVarHost)
	if err != nil {
		log := logrus.WithFields(logrus.Fields{
			"module": "main",
			"key":    "host",
		})
		log.Fatalf("No value set, please pass -host or set %s", envVarHost)
	}

	port, err := argOrEnvVar(*Port, envVarPort)
	if err != nil {
		log := logrus.WithFields(logrus.Fields{
			"module": "main",
			"key":    "port",
		})
		log.Fatalf("No value set, please pass -port or set %s", envVarPort)
	}

	token, err := argOrEnvVar(*Token, envVarToken)
	if err != nil {
		log := logrus.WithFields(logrus.Fields{
			"module": "main",
			"key":    "token",
		})
		log.Fatalf("No value set, please pass -token or set %s", envVarToken)
	}

	uri := fmt.Sprintf("%s:%d", host, port)

	httptransport.DefaultTimeout = 300 * time.Second
	transport := httptransport.New(uri, "/v1/", nil)

	apiToken := httptransport.APIKeyAuth("X-Access-Token", "header", token.(string))

	transport.Consumers["application/json"] = runtime.JSONConsumer()
	transport.Consumers["application/vnd.api+json"] = runtime.JSONConsumer()
	client := apiclient.New(transport, strfmt.Default)

	cfg, err := config.New(*ConfigFile)
	if err != nil {
		log.Fatalf("config.New: %v", err)
	}

	ui, err := webui.NewWebUi(cfg.WebUi, apiToken, client)

	err = ui.Run()
	if err != nil {
		log.Fatalf("ui.Run: %v", err)
	}
}
