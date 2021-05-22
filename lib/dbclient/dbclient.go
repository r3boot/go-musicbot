package dbclient

import (
	"fmt"
	"log"
	"os"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/r3boot/test/lib/config"
)

/*
TODO: create extension pg_trgm;
*/

const (
	T_BLACKLIST = "blacklist"
)

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

	log := log.New(os.Stdout, "go-pg: ", log.Lshortfile)
	pg.SetLogger(log)

	err := client.Connect()
	if err != nil {
		return nil, fmt.Errorf("localDbClient.New: %v", err)
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
		return fmt.Errorf("DbClient.Connect: Failed to connect to database")
	}
	// o.log.Debugf("DbClient.Connect: Connected to pg://%s:***@%s/%s", o.cfg.Username, o.cfg.Address, o.cfg.Database)

	for _, model := range []interface{}{&Track{}} {
		err := o.db.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return fmt.Errorf("DbClient.Connect db.CreateTable: %v", err)
		}

	}

	return nil
}
