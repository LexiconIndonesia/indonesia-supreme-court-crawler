package scrapper

import (
	"lexicon/indonesia-supreme-court-crawler/common"
	crawler_service "lexicon/indonesia-supreme-court-crawler/crawler/services"
	"lexicon/indonesia-supreme-court-crawler/scrapper/models"
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

	var result []models.Metadata

	scrapper, err := buildScrapper(&result)

	if err != nil {
		log.Error().Err(err).Msg("Error building scrapper")
	}

	for _, url := range list {
		// scrape url
		scrapper.Visit(url.Url)
		// save to db
		// update status

	}

}

func buildScrapper(metadata *[]models.Metadata) (*colly.Collector, error) {
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

		*metadata = append(*metadata, models.Metadata{
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
		})

		// model = models.Metadata{
		// 	Title: title,
		// 	Defendant: string(defendant),
		// }

	})
	c.OnRequest(func(r *colly.Request) {
		log.Info().Msg("Visiting: " + r.URL.String())
	})
	return c, nil
}

func findValue(e *colly.HTMLElement, selector string) string {
	var value string
	e.DOM.Find("td").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), selector) {
			value = s.Next().Text()
		}
	})
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(value), "\n", " "), "—", "")
}
