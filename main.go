package main

import (
	"context"
	"lexicon/indonesia-supreme-court-crawler/common"
	"lexicon/indonesia-supreme-court-crawler/crawler"
	"lexicon/indonesia-supreme-court-crawler/scrapper"

	"github.com/golang-module/carbon/v2"

	"github.com/rs/zerolog/log"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func main() {
	// INITIATE CONFIGURATION
	err := godotenv.Load()
	if err != nil {
		log.Error().Err(err).Msg("Error loading .env file")
	}
	cfg := defaultConfig()
	cfg.loadFromEnv()
	carbon.SetDefault(carbon.Default{
		Layout:       carbon.ISO8601Layout,
		Timezone:     carbon.UTC,
		WeekStartsAt: carbon.Monday,
		Locale:       "en",
	})

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
	rootCommand := &cobra.Command{
		Use:   "indonesia-supreme-court-crawler",
		Short: "Crawl Supreme Court of Indonesia",
		Long:  "Crawl Supreme Court of Indonesia",
		Run: func(cmd *cobra.Command, args []string) {

			log.Info().Msg("Start Crawler")
			crawler.StartCrawler()
			log.Info().Msg("Start Scrapper")
			scrapper.StartScraper()
			log.Info().Msg("Finish")
		},
	}

	rootCommand.AddCommand(crawlerCommand())
	rootCommand.AddCommand(scrapperCommand())

	rootCommand.Execute()

}

func crawlerCommand() *cobra.Command {

	return &cobra.Command{
		Use:   "crawler",
		Short: "Crawl Supreme Court of Indonesia",
		Long:  "Crawl Supreme Court of Indonesia",
		Run: func(cmd *cobra.Command, args []string) {

			crawler.StartCrawler()
		},
	}
}

func scrapperCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "scrapper",
		Short: "Scrapper Supreme Court of Indonesia",
		Long:  "Scrapper Supreme Court of Indonesia",
		Run: func(cmd *cobra.Command, args []string) {

			scrapper.StartScraper()
		},
	}
}
