package main

import (
	"context"
	"lexicon/indonesia-supreme-court-crawler/common"
	"lexicon/indonesia-supreme-court-crawler/scrapper"

	"github.com/golang-module/carbon/v2"

	"github.com/rs/zerolog/log"

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

	ctx := context.Background()

	carbon.SetDefault(carbon.Default{
		Layout:       carbon.ISO8601Layout,
		Timezone:     carbon.UTC,
		WeekStartsAt: carbon.Monday,
		Locale:       "en",
	})

	// INITIATE DATABASES
	// PGSQL
	pgsqlClient, err := pgxpool.New(ctx, cfg.PgSql.ConnStr())

	if err != nil {
		log.Error().Err(err).Msg("Unable to connect to PGSQL Database")
	}
	defer pgsqlClient.Close()

	err = common.SetDatabase(pgsqlClient)
	if err != nil {
		log.Error().Err(err).Msg("Unable to set database")
	}

	// GCS
	// gcsClient, err := storage.NewClient(ctx)
	// if err != nil {
	// 	log.Error().Err(err).Msg("Unable to connect to GCS")
	// }

	// defer gcsClient.Close()

	// err = common.SetStorageClient(gcsClient)
	if err != nil {
		log.Error().Err(err).Msg("Unable to set storage client")

	}

	// Start Crawler

	// crawler.StartCrawler()

	// Start Scrapper

	scrapper.StartScraper()

}
