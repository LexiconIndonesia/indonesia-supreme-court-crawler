package main

import (
	"context"
	"lexicon/indonesia-supreme-court-crawler/common"
	"lexicon/indonesia-supreme-court-crawler/repository"

	"github.com/rs/zerolog/log"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// INITIATE CONFIGURATION
	err := godotenv.Load()
	if err != nil {
		log.Error().Err(err).Msg("Error loading .env file")
	}
	cfg := defaultConfig()
	cfg.loadFromEnv()

	// INITIATE DATABASES
	// PGSQL
	ctx := context.Background()

	pgsqlClient, err := pgxpool.New(ctx, cfg.PgSql.ConnStr())

	if err != nil {
		log.Error().Err(err).Msg("Unable to connect to PGSQL Database")
	}
	defer pgsqlClient.Close()

	err = common.SetDatabase(pgsqlClient)
	if err != nil {
		log.Error().Err(err).Msg("Unable to set database")
	}

	query := repository.New(pgsqlClient)

	err = common.SetQuery(query)
	if err != nil {
		log.Error().Err(err).Msg("Unable to set query")
	}

	// GCS
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Unable to connect to GCS")
	}

	defer gcsClient.Close()

	err = common.SetStorageClient(gcsClient)
	if err != nil {
		log.Error().Err(err).Msg("Unable to set storage client")

	}

}
