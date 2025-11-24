package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kratos069/message-app/api"
	"github.com/kratos069/message-app/db/sqlc/db-gen"
	"github.com/kratos069/message-app/mail"
	"github.com/kratos069/message-app/util"
	"github.com/kratos069/message-app/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

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

	// Create context that listens for interrupt signals
	ctx, stop := signal.NotifyContext(
		context.Background(), interruptSignals...)
	defer stop()

	// conn to database
	connPool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal().Msg("cannot connect to the db")
	}

	// run db migrations
	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(connPool)

	// redis options
	redisOpts := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpts)

	// run Redis Task Processor
	waitGroup, ctx := errgroup.WithContext(ctx)
	go runTaskProcessor(ctx, waitGroup, config, redisOpts, store)

	// Start main Gin server & debug server
	ginServer := runGinServer(config, store, taskDistributor)
	debugServer := runDebugServer(config)

	// Wait for interrupt signal
	<-ctx.Done()

	// Restore default behavior on the interrupt signal
	stop()
	log.Info().Msg("Shutting down gracefully, press Ctrl+C again to force")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown services in order
	log.Info().Msg("Shutting down HTTP server...")
	if err := ginServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server forced to shutdown")
	} else {
		log.Info().Msg("HTTP server stopped gracefully")
	}

	// Shutdown debug server
	if debugServer != nil {
		log.Info().Msg("Shutting down debug server...")
		if err := debugServer.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Debug server forced to shutdown")
		} else {
			log.Info().Msg("Debug server stopped gracefully")
		}
	}

	// Close database connection pool
	log.Info().Msg("Closing database connections...")
	connPool.Close()
	log.Info().Msg("Database connections closed")
	log.Info().Msg("Server exited successfully")
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

// run Main Gin Server
func runGinServer(config util.Config, store db.Store,
	taskDistributor worker.TaskDistributor) *api.Server {
	// Start main Gin server
	ginServer, err := api.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("cannot create gin server:")
	}

	// Start Gin HTTP server in a goroutine
	go func() {
		log.Info().Msgf("Starting HTTP server on %s", config.HTTPServerAddress)
		if err := ginServer.Start(
			config.HTTPServerAddress); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	return ginServer
}

// run Debug Server and return the server instance for graceful shutdown
func runDebugServer(config util.Config) *http.Server {
	debugServer := &http.Server{
		Addr:    config.DebugHost,
		Handler: util.Mux(),
	}

	go func() {
		log.Info().Msgf("Starting debug server on %s", config.DebugHost)
		if err := debugServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Debug server error")
		}
	}()

	return debugServer
}

func runTaskProcessor(ctx context.Context, waitGroup *errgroup.Group,
	config util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)

	log.Info().Msg("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown task processor")

		// taskProcessor.Shutdown()
		// log.Info().Msg("task processor is stopped")

		return nil
	})
}
