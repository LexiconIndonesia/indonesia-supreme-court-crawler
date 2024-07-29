package models

import (
	"github.com/golang-module/carbon/v2"
	"github.com/guregu/null"
	"github.com/oklog/ulid"
)

type Extraction struct {
	Id            ulid.ULID
	UrlFrontierId ulid.ULID
	SiteContent   null.String
	ArtifactLink  null.String
	RawPageLink   null.String
	CreatedAt     carbon.DateTime
	UpdatedAt     carbon.DateTime
	Language      string
}
