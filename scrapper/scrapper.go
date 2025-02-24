package scrapper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"lexicon/indonesia-supreme-court-crawler/common"
	crawler_model "lexicon/indonesia-supreme-court-crawler/crawler/models"
	crawler_service "lexicon/indonesia-supreme-court-crawler/crawler/services"
	"lexicon/indonesia-supreme-court-crawler/repository"
	"lexicon/indonesia-supreme-court-crawler/scrapper/models"
	"lexicon/indonesia-supreme-court-crawler/scrapper/services"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func StartScraper() {
	// Fetch unscrapped url frontier from db
	list, err := crawler_service.GetUnscrappedUrlFrontiers(context.Background(), 100)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching unscrapped url frontier")
	}

	log.Info().Msg("Unscrapped URLs: " + strconv.Itoa(len(list)))

	q, err := queue.New(1, &queue.InMemoryQueueStorage{MaxSize: 100000})

	if err != nil {
		log.Error().Err(err).Msg("Error creating queue")
	}

	scrapper, err := buildScrapper()
	if err != nil {
		log.Error().Err(err).Msg("Error building scrapper")
	}

	for _, url := range list {
		// scrape url
		q.AddURL(url.Url)
	}
	q.Run(scrapper)
	scrapper.Wait()
}

func buildScrapper() (*colly.Collector, error) {
	newExtraction := repository.Extraction{}
	var newMetadata models.ExtractionMetadata
	defendantFirstLayerRegex, err := regexp.Compile(`(?mi).*—\s`)
	if err != nil {
		log.Error().Err(err).Msg("Error compiling regex")
		return nil, err
	}
	defendantRegex, err := regexp.Compile(`(?mi)Putusan\s[A-Z0-9a-z\/\.\s\-]*.\sPembanding/Penuntut\sUmum\s\:\s[A-Z\,\s]*Terbanding/Terdakwa\s\:\s`)
	if err != nil {
		log.Error().Err(err).Msg("Error compiling regex")
		return nil, err
	}
	secondDefendantRegex, err := regexp.Compile(`^(.*):\s?(.*)$`)
	if err != nil {
		log.Error().Err(err).Msg("Error compiling regex")
		return nil, err
	}

	c := colly.NewCollector(
		colly.AllowedDomains(common.CRAWLER_DOMAIN),
		// colly.Async(true),
		colly.MaxDepth(1),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  common.CRAWLER_DOMAIN,
		Delay:       time.Second * 2,
		RandomDelay: time.Second * 2,
	})

	c.SetRequestTimeout(time.Minute * 2)

	c.OnHTML("table.table", func(e *colly.HTMLElement) {
		title := e.ChildText("h2")
		defendantClean := defendantFirstLayerRegex.ReplaceAll([]byte(title), []byte(""))
		defendant := defendantRegex.ReplaceAll([]byte(defendantClean), []byte(""))
		usedDefendant := string(defendant)
		if match := secondDefendantRegex.FindStringSubmatch(string(usedDefendant)); match != nil {
			usedDefendant = match[2]
		}
		id := sha256.Sum256([]byte(title))
		newMetadata = models.ExtractionMetadata{
			Id:                       hex.EncodeToString(id[:]),
			Title:                    title,
			Defendant:                string(usedDefendant),
			Number:                   findValue(e, "Nomor"),
			ProcessLevel:             findValue(e, "Tingkat Proses"),
			Classification:           findValue(e, "Klasifikasi"),
			Keywords:                 findValue(e, "Kata Kunci"),
			Year:                     findValue(e, "Tahun"),
			RegistrationDate:         findValue(e, "Tanggal Register"),
			JudicalInstitution:       findValue(e, "Lembaga Peradilan"),
			TypeOfJudicalInstitution: findValue(e, "Jenis Lembaga Peradilan"),
			ChiefJustice:             findValue(e, "Hakim Ketua"),
			MemberJudge:              findValue(e, "Hakim Anggota"),
			Clerk:                    findValue(e, "Panitera"),
			Verdict:                  findValue(e, "Amar"),
			OtherVerdict:             findValue(e, "Amar Lainnya"),
			VerdictNote:              findValue(e, "Catatan Amar"),
			CourtDate:                findValue(e, "Tanggal Musyawarah"),
			AnnouncementDate:         findValue(e, "Tanggal Dibacakan"),
			Rule:                     findValue(e, "Kaidah"),
			Abstract:                 findValue(e, "Abstrak"),
		}

		newExtraction.SiteContent = &e.Text

		// model = models.Metadata{
		// 	Title: title,
		// 	Defendant: string(defendant),
		// }
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		var pdfUrl string
		if strings.Contains(url, "pdf") {
			pdfUrl = url
		}

		if pdfUrl != "" {
			log.Info().Msg("Found PDF: " + pdfUrl)
			newMetadata.PdfUrl = pdfUrl

		}
	})
	c.OnRequest(func(r *colly.Request) {
		log.Info().Msg("Visiting: " + r.URL.String())
	})
	c.OnScraped(func(r *colly.Response) {
		requestUrl := r.Request.URL.String()
		frontierId := sha256.Sum256([]byte(requestUrl))
		newExtraction.UrlFrontierID = hex.EncodeToString(frontierId[:])
		newExtraction.ID = hex.EncodeToString(frontierId[:])
		newExtraction.RawPageLink = &requestUrl
		newExtraction.Metadata = newMetadata

		if newExtraction.Metadata.PdfUrl != "" {
			log.Info().Msg("Uploading PDF to GCS: " + newExtraction.Metadata.PdfUrl)

			artifact, err := services.HandlePdf(context.Background(), newExtraction.Metadata.Id+".pdf", newExtraction.Metadata.PdfUrl)
			if err != nil {
				log.Error().Err(err).Msg("Error handling pdf")
			}

			newExtraction.ArtifactLink = &artifact.B
		}
		log.Info().Msg("Upserting extraction: " + newExtraction.ID)
		err = services.UpsertExtraction(context.Background(), []repository.Extraction{newExtraction})
		if err != nil {
			log.Error().Err(err).Msg("Error upserting extraction")
		}
		log.Info().Msg("Updating url frontier status: " + newExtraction.UrlFrontierID)
		err = crawler_service.UpdateFrontierStatuses(context.Background(), []lo.Tuple2[string, int16]{{newExtraction.UrlFrontierID, int16(crawler_model.URL_FRONTIER_STATUS_CRAWLED)}})
		if err != nil {
			log.Error().Err(err).Msg("Error updating url frontier status")
		}

		log.Info().Msg("Scraped: " + r.Request.URL.String())
	})

	return c, nil
}

func findValue(e *colly.HTMLElement, selector string) string {
	var value string
	e.DOM.Find("td.text-right").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), selector) {
			value = s.Next().Text()
		}
	})

	value = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(value), "\n", " "), "—", "")
	if selector == "Klasifikasi" {
		// 1. Remove duplicate \t
		re := regexp.MustCompile(`\t+`)
		singleTab := re.ReplaceAllString(value, "\t")

		// 2. Change \t to hyphen
		hyphenated := strings.ReplaceAll(singleTab, "\t", "-")

		// 3. Remove duplicate spaces
		reSpace := regexp.MustCompile(`\s+`)
		value = reSpace.ReplaceAllString(hyphenated, " ")

	}

	return value
}
