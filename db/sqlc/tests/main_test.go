package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kratos069/message-app/db/sqlc/db-gen"
	"github.com/kratos069/message-app/util"
)

var testStore db.Store

func TestMain(m *testing.M) {
	var err error

	config, err := util.LoadConfig("../../..")
	if err != nil {
		log.Fatalln("cannot load config file:", err)
	}

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatalln("cannot connect to the db:", err)
	}

	testStore = db.NewStore(connPool)

	os.Exit(m.Run())
}
