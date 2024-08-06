package crawler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"lexicon/indonesia-supreme-court-crawler/common"
	"lexicon/indonesia-supreme-court-crawler/crawler/models"
	"lexicon/indonesia-supreme-court-crawler/crawler/services"
	"regexp"
	"strconv"
	"time"

	"github.com/golang-module/carbon/v2"

	"github.com/gocolly/colly/v2"
	"github.com/rs/zerolog/log"
)

func StartCrawler() {
	c := colly.NewCollector(
		colly.AllowedDomains("putusan3.mahkamahagung.go.id"),
	)
	startPage := 1

	url := fmt.Sprintf("https://putusan3.mahkamahagung.go.id/search.html?q=korupsi&page=%d&obf=TANGGAL_PUTUS&obm=desc", startPage)

	lastPage := getLastPage(url)

	totalData := (20 * lastPage) - ((startPage - 1) * 20)
	log.Info().Msg("Total Data: " + strconv.Itoa(totalData))
	for i := startPage; i <= lastPage; i++ {
		details := crawlUrl(c, fmt.Sprintf("https://putusan3.mahkamahagung.go.id/search.html?q=korupsi&page=%d&obf=TANGGAL_PUTUS&obm=desc", i))

		log.Info().Msg("Details: " + strconv.Itoa(len(details)))
		allDetails := []models.UrlFrontier{}

		for _, detail := range details {

			id := sha256.Sum256([]byte(detail))
			currentTime := carbon.Now().ToDateTimeStruct()
			allDetails = append(allDetails, models.UrlFrontier{
				Id:        hex.EncodeToString(id[:]),
				Url:       detail,
				Domain:    common.CRAWLER_DOMAIN,
				Crawler:   common.CRAWLER_NAME,
				Status:    models.URL_FRONTIER_STATUS_NEW,
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
			})
		}

		err := services.UpsertUrl(allDetails)
		if err != nil {
			log.Error().Err(err).Msg("Error upserting url")
		}

		time.Sleep(time.Second * 2)
	}

}

func getLastPage(url string) int {
	lastPage := 0
	c := colly.NewCollector(
		colly.AllowedDomains("putusan3.mahkamahagung.go.id"),
	)

	// find lastPage links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		var err error
		linkClass := e.Attr("class")
		isLast := e.Text == "Last"
		if linkClass != "" && isNavigationLink(linkClass) && isLast {
			log.Info().Msg("Found Navigation Link: " + linkClass)
			lastPage, err = strconv.Atoi(e.Attr("data-ci-pagination-page"))
			if err != nil {
				log.Error().Err(err).Msg("Error parsing page number")
			}
		}

	})

	c.OnRequest(func(r *colly.Request) {
		log.Info().Msg("Visiting: " + r.URL.String())
	})
	c.Visit(url)
	return lastPage

}
func crawlUrl(c *colly.Collector, url string) []string {
	log.Info().Msg("Crawling URL: " + url)

	var detailUrls []string
	// find current page links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if link != "" && isDetailPage(link) {
			log.Info().Msg("Found Detail Page: " + link)
			detailUrls = append(detailUrls, link)
		}

	})

	c.OnRequest(func(r *colly.Request) {
		log.Info().Msg("Visiting: " + r.URL.String())
	})

	c.Visit(url)

	return detailUrls
	//find pagination links

}

func isDetailPage(link string) bool {
	checker, err := regexp.Compile("/direktori/putusan")
	if err != nil {
		log.Error().Err(err).Msg("Regex Compile Error")
		return false
	}

	return checker.MatchString(link)
}

func isNavigationLink(linkClass string) bool {
	return linkClass == "page-link"
}
