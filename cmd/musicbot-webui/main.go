package main

import (
	"flag"
	"fmt"
	"github.com/r3boot/go-musicbot/lib/log"
	"github.com/r3boot/go-musicbot/lib/utils"
	"os"
	"strconv"
	"time"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/webui"

	"github.com/go-openapi/runtime"

	"github.com/go-openapi/strfmt"

	httptransport "github.com/go-openapi/runtime/client"

	"github.com/r3boot/go-musicbot/lib/apiclient"
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
	)
	flag.Parse()

	log.NewLogger(*LogLevel, *LogJson)

	cfg, err := config.New(*ConfigFile)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "main",
			"function": "main",
			"call":     "config.New",
			"cfgfile":  *ConfigFile,
		}, err.Error())
	}

	host, err := utils.ArgOrEnvVar(*Host, envVarHost, cfg.WebApi.Address)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "main",
			"function": "main",
			"call":     "utils.ArgOrEnvVar",
		}, "no value set, please pass -host or set %s", envVarHost)
	}

	port, err := utils.ArgOrEnvVar(*Port, envVarPort, strconv.Itoa(cfg.WebApi.Port))
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "main",
			"function": "main",
			"call":     "utils.ArgOrEnvVar",
		}, "no value set, please pass -port or set %s", envVarPort)
	}

	token, err := utils.ArgOrEnvVar(*Token, envVarToken, cfg.WebUi.Token)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "main",
			"function": "main",
			"call":     "utils.ArgOrEnvVar",
		}, "no value set, please pass -token or set %s", envVarToken)
	}

	uri := fmt.Sprintf("%s:%d", host, port)

	httptransport.DefaultTimeout = 300 * time.Second
	transport := httptransport.New(uri, "/v1/", nil)

	apiToken := httptransport.APIKeyAuth("X-Access-Token", "header", token.(string))

	transport.Consumers["application/json"] = runtime.JSONConsumer()
	transport.Consumers["application/vnd.api+json"] = runtime.JSONConsumer()
	client := apiclient.New(transport, strfmt.Default)

	ui, err := webui.NewWebUi(cfg.WebUi, apiToken, client)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "main",
			"function": "main",
			"call":     "webui.NewWebUi",
		}, err.Error())
	}

	err = ui.Run()
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "main",
			"function": "main",
			"call":     "ui.Run",
		}, err.Error())
	}
}
