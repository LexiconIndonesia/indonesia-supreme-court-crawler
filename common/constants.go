package common

import "fmt"

const (
	CRAWLER_NAME   string = "indonesia-supreme-court-crawler"
	CRAWLER_DOMAIN string = "putusan3.mahkamahagung.go.id"
	GCS_BUCKET     string = "lexicon-bo-bucket"
)

var (
	GCS_FOLDER      string = fmt.Sprintf("%s/%s", CRAWLER_NAME, "judgements")
	GCS_HTML_FOLDER string = fmt.Sprintf("%s/%s", CRAWLER_NAME, "html")
)
