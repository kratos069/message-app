package main

import (
	"context"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kratos069/message-app/api"
	"github.com/kratos069/message-app/db/sqlc/db-gen"
	"github.com/kratos069/message-app/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// load ENV variables
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msg("cannot load config file:")
	}

	// pretty logs in dev (instead of json)
	if config.Environment == "development" {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// conn to database
	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Msg("cannot connect to the db")
	}

	// run db migrations
	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(connPool)

	runGinServer(config, store)
	// go runGatewayServer(config, store)
	//runGrpcServer(config, store)
}

func runDBMigration(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create new migration instance:")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("failed to run migrate up:")
	}

	log.Info().Msg("db migrated successfully")
}

// run Gin Server for HTTP requests
func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create gin server:")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot start server:")
	}
}
