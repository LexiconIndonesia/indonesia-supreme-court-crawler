package scrapper

import (
	"crypto/sha256"
	"encoding/hex"
	"lexicon/indonesia-supreme-court-crawler/common"
	crawler_service "lexicon/indonesia-supreme-court-crawler/crawler/services"
	"lexicon/indonesia-supreme-court-crawler/scrapper/models"
	"lexicon/indonesia-supreme-court-crawler/scrapper/services"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"

	"github.com/rs/zerolog/log"
)

func StartScraper() {

	// Fetch unscrapped url frontier from db
	list, err := crawler_service.GetUnscrapedUrlFrontier()
	if err != nil {
		log.Error().Err(err).Msg("Error fetching unscrapped url frontier")
	}

	log.Info().Msg("Unscrapped URLs: " + strconv.Itoa(len(list)))

	var result []models.Extraction

	scrapper, err := buildScrapper(&result)

	if err != nil {
		log.Error().Err(err).Msg("Error building scrapper")
	}

	scrapper.Visit(list[0].Url)
	scrapper.Visit(list[1].Url)

	// for _, url := range list {
	// 	// scrape url
	// 	scrapper.Visit(url.Url)

	// }

	for _, datum := range result {
		log.Info().Msg("Uploading PDF to GCS: " + datum.Metadata.PdfUrl)

		// artifact, err := services.HandlePdf(datum.Metadata, datum.Metadata.PdfUrl, datum.Metadata.Id+".pdf")
		// if err != nil {
		// 	log.Error().Err(err).Msg("Error handling pdf")
		// }

		// datum.AddArtifactLink(artifact)
		err = services.UpsertExtraction(datum)
		if err != nil {
			log.Error().Err(err).Msg("Error upserting extraction")
		}
		time.Sleep(time.Second * 2)
	}

}

func buildScrapper(extraction *[]models.Extraction) (*colly.Collector, error) {
	var newExtraction = models.NewExtraction()
	var newMetadata models.Metadata
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

	c := colly.NewCollector(
		colly.AllowedDomains(common.CRAWLER_DOMAIN),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  common.CRAWLER_DOMAIN,
		Parallelism: 10,
		Delay:       time.Second * 2,
		RandomDelay: time.Second * 2,
	})

	c.OnHTML("table.table", func(e *colly.HTMLElement) {
		title := e.ChildText("h2")
		defendantClean := defendantFirstLayerRegex.ReplaceAll([]byte(title), []byte(""))

		defendant := defendantRegex.ReplaceAll([]byte(defendantClean), []byte(""))
		id := sha256.Sum256([]byte(title))
		newMetadata = models.Metadata{
			Id:                       hex.EncodeToString(id[:]),
			Title:                    title,
			Defendant:                string(defendant),
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

		newExtraction.AddSiteContent(e.Text)

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
		newExtraction.AddRawPageLink(r.URL.String())
		log.Info().Msg("Visiting: " + r.URL.String())
	})
	c.OnScraped(func(r *colly.Response) {
		frontierId := sha256.Sum256([]byte(r.Request.URL.String()))
		newExtraction.UrlFrontierId = hex.EncodeToString(frontierId[:])
		newExtraction.Id = hex.EncodeToString(frontierId[:])
		log.Info().Msg("Scraped: " + r.Request.URL.String())
		newExtraction.AddMetadata(newMetadata)
		*extraction = append(*extraction, newExtraction)
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
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(value), "\n", " "), "—", "")
}
