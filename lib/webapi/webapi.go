package webapi

import (
	"fmt"
	"github.com/go-openapi/loads"
	"github.com/r3boot/go-musicbot/lib/apiserver"
	"github.com/r3boot/go-musicbot/lib/apiserver/operations"
	"github.com/r3boot/go-musicbot/lib/config"
)

type WebApi struct {
	allCfg *config.Config
	cfg    *config.WebApi
}

func NewWebApi(cfg *config.Config) (*WebApi, error) {
	api := &WebApi{
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

	apiserver.Config = wa.allCfg

	server.ConfigureAPI()
	server.Host = host
	server.Port = port

	if err := server.Serve(); err != nil {
		return fmt.Errorf("Serve: %v", err)
	}

	return nil
}
