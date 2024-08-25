package models

import (
	"context"
	"encoding/json"

	"github.com/golang-module/carbon/v2"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v5"
)

type Extraction struct {
	Id            string
	UrlFrontierId string
	SiteContent   null.String
	ArtifactLink  null.String
	RawPageLink   null.String
	Metadata      Metadata
	CreatedAt     carbon.DateTime
	UpdatedAt     carbon.DateTime
	Language      string
}

func (e *Extraction) AddSiteContent(content string) {
	e.SiteContent = null.StringFrom(content)
}

func (e *Extraction) AddArtifactLink(link string) {
	e.ArtifactLink = null.StringFrom(link)
}

func (e *Extraction) AddRawPageLink(link string) {
	e.RawPageLink = null.StringFrom(link)
}

func (e *Extraction) AddMetadata(metadata Metadata) {
	e.Metadata = metadata
}
func (e *Extraction) UpdateUpdatedAt() {
	e.UpdatedAt = carbon.Now().ToDateTimeStruct()
}
func NewExtraction() Extraction {
	return Extraction{
		Id:            "",
		UrlFrontierId: "",
		SiteContent:   null.StringFrom(""),
		ArtifactLink:  null.StringFrom(""),
		RawPageLink:   null.StringFrom(""),
		Metadata:      Metadata{},
		CreatedAt:     carbon.Now().ToDateTimeStruct(),
		UpdatedAt:     carbon.Now().ToDateTimeStruct(),
		Language:      "id",
	}
}

func UpsertExtraction(ctx context.Context, tx pgx.Tx, extraction Extraction) error {

	sql := `INSERT INTO extraction (id, url_frontier_id, site_content, artifact_link, raw_page_link, metadata, created_at, updated_at, language) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (id) DO UPDATE SET site_content = EXCLUDED.site_content, artifact_link = EXCLUDED.artifact_link, raw_page_link = EXCLUDED.raw_page_link, metadata = EXCLUDED.metadata, updated_at = EXCLUDED.updated_at`

	metadataJson, err := json.Marshal(extraction.Metadata)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, sql, extraction.Id, extraction.UrlFrontierId, extraction.SiteContent, extraction.ArtifactLink, extraction.RawPageLink, metadataJson, extraction.CreatedAt.ToIso8601String(), extraction.UpdatedAt.ToIso8601String(), extraction.Language)
	if err != nil {
		return err
	}

	return nil
}
