package dbclient

import (
	"fmt"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/log"
)

/*
TODO: migrate to pg/v10
TODO: create extension pg_trgm;
*/

type DbClient struct {
	cfg *config.PostgresConfig
	db  *pg.DB
}

var (
	client *DbClient
)

func NewDbClient(cfg *config.PostgresConfig) (*DbClient, error) {
	client = &DbClient{
		cfg: cfg,
	}

	/* TODO
	log := log.New(os.Stdout, "go-pg: ", log.Lshortfile)
	pg.SetLogger(log.GetLogger())
	*/

	err := client.Connect()
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "dbclient",
			"function": "NewDbClient",
			"call":     "client.Connect",
		}, err.Error())
		return nil, fmt.Errorf("DbClient.New: %v", err)
	}

	return client, nil
}

func (o *DbClient) Connect() error {
	client.db = pg.Connect(&pg.Options{
		Addr:     o.cfg.Address,
		User:     o.cfg.Username,
		Password: o.cfg.Password,
		Database: o.cfg.Database,
	})

	if o.db == nil {
		log.Fatalf(log.Fields{
			"package":  "dbclient",
			"function": "Connect",
			"call":     "pg.Connect",
			"address":  o.cfg.Address,
			"username": o.cfg.Username,
			"database": o.cfg.Database,
		}, "failed to connect")
		return fmt.Errorf("failed to connect")
	}
	log.Debugf(log.Fields{
		"package":  "dbclient",
		"function": "Connect",
		"address":  o.cfg.Address,
		"username": o.cfg.Username,
		"database": o.cfg.Database,
	}, "connected to database")

	models := []interface{}{
		&Track{},
	}
	for _, model := range models {
		err := o.db.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			log.Fatalf(log.Fields{
				"package":  "dbclient",
				"function": "Connect",
				"call":     "o.db.CreateTable",
				"address":  o.cfg.Address,
				"username": o.cfg.Username,
				"database": o.cfg.Database,
			}, err.Error())
			return fmt.Errorf("failed to create table")
		}
	}

	return nil
}
