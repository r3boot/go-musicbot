package webapi

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/go-openapi/loads"
	"github.com/jessevdk/go-flags"
	"github.com/r3boot/test/lib/apiserver"
	"github.com/r3boot/test/lib/apiserver/operations"
	"github.com/r3boot/test/lib/config"
)

type WebApi struct {
	log    *logrus.Entry
	allCfg *config.Config
	cfg    *config.WebApi
}

func NewWebApi(cfg *config.Config) (*WebApi, error) {
	api := &WebApi{
		log: logrus.WithFields(logrus.Fields{
			"caller": "WebApi",
		}),
		allCfg: cfg,
		cfg:    cfg.WebApi,
	}

	return api, nil
}

func (wa *WebApi) Run(host string, port int) error {
	swaggerSpec, err := loads.Embedded(apiserver.SwaggerJSON, apiserver.FlatSwaggerJSON)
	if err != nil {
		return fmt.Errorf("Embedded: %v", err)
	}

	api := operations.NewMusicbotAPI(swaggerSpec)
	server := apiserver.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "musicbot api"
	parser.LongDescription = "The api serving the musicbot functionality."

	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			return fmt.Errorf("AddGroup: %v", err)
		}
	}

	apiserver.Config = wa.allCfg

	parser.Parse()
	server.ConfigureAPI()
	server.Host = host
	server.Port = port

	if err := server.Serve(); err != nil {
		return fmt.Errorf("Serve: %v", err)
	}

	return nil
}
