package main

import (
	"context"
	"fmt"
	"lexicon/indonesia-supreme-court-crawler/common"
	"lexicon/indonesia-supreme-court-crawler/crawler"

	"github.com/golang-module/carbon/v2"

	"github.com/rs/zerolog/log"

	"github.com/gocolly/colly/v2"
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

	common.SetDatabase(pgsqlClient)

	// Start Crawler
	c := colly.NewCollector(
		colly.AllowedDomains("putusan3.mahkamahagung.go.id"),
	)
	startPage := 1

	startUrl := fmt.Sprintf("https://putusan3.mahkamahagung.go.id/search.html?q=korupsi&page=%d&obf=TANGGAL_PUTUS&obm=desc", startPage)
	crawler.StartCrawler(c, startUrl, startPage)

	// Start Scrapper
}
