package main

import (
	"flag"
	"fmt"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/r3boot/go-musicbot/lib/apiclient"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/ircbot"
	"github.com/r3boot/go-musicbot/lib/log"
	"github.com/r3boot/go-musicbot/lib/utils"
	"strconv"
	"time"
)

const (
	defConfigFileValue = "/etc/musicbot.yml"
	defHostValue       = "localhost"
	defPortValue       = 8080

	helpHost       = "Host to connect to"
	helpPort       = "Port to connect to"
	helpConfigFile = "Path to the configuration file"

	musicbotSubmitter = "musicbot"

	envVarPrefix = "MUSICBOT_IRCBOT_"
	envVarHost   = envVarPrefix + "HOST"
	envVarPort   = envVarPrefix + "PORT"
	envVarToken  = envVarPrefix + "TOKEN"
)

func main() {
	var (
		LogLevel   = flag.String("loglevel", "INFO", "Log level to use (INFO, DEBUG)")
		LogJson    = flag.Bool("json", false, "Output logging in JSON format")
		ConfigFile = flag.String("config", defConfigFileValue, helpConfigFile)

		Host  = flag.String("host", defHostValue, helpHost)
		Port  = flag.Int("port", defPortValue, helpPort)
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

	token, err := utils.ArgOrEnvVar(*Token, envVarToken, cfg.IrcBot.Token)
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

	bot, err := ircbot.NewIrcBot(cfg, client, apiToken, &ircbot.IrcBotParams{
		Nickname:  cfg.IrcBot.Nickname,
		Server:    cfg.IrcBot.Server,
		Port:      cfg.IrcBot.Port,
		Channel:   cfg.IrcBot.Channel,
		Tls:       cfg.IrcBot.Tls,
		VerifyTls: cfg.IrcBot.VerifyTls,
	})
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "main",
			"function": "main",
			"call":     "ircbot.NewIrcBot",
		}, err.Error())
	}

	err = bot.Run()
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "main",
			"function": "main",
			"call":     "bot.Run",
		}, err.Error())
	}
}
