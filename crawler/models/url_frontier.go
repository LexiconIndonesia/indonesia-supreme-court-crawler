package models

import (
	"context"
	"lexicon/indonesia-supreme-court-crawler/common"

	"github.com/golang-module/carbon/v2"
	"github.com/jackc/pgx/v5"
)

const (
	URL_FRONTIER_STATUS_NEW     = 0
	URL_FRONTIER_STATUS_CRAWLED = 1
	URL_FRONTIER_STATUS_ERROR   = 2
)

type UrlFrontier struct {
	Id        string
	Domain    string
	Url       string
	Crawler   string
	Status    uint8
	CreatedAt carbon.DateTime
	UpdatedAt carbon.DateTime
}

func UpsertUrlFrontier(ctx context.Context, tx pgx.Tx, urlFrontier []UrlFrontier) error {

	sql := `INSERT INTO url_frontier (id, domain, url, crawler, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (id) DO UPDATE SET domain = EXCLUDED.domain, url = EXCLUDED.url, crawler = EXCLUDED.crawler, status = EXCLUDED.status, updated_at = EXCLUDED.updated_at`

	batch := &pgx.Batch{}

	for _, url := range urlFrontier {
		batch.Queue(sql, url.Id, url.Domain, url.Url, url.Crawler, url.Status, url.CreatedAt.ToIso8601String(), url.UpdatedAt.ToIso8601String())

	}
	res := tx.SendBatch(ctx, batch)

	return res.Close()
}

func UpdateUrlFrontiersStatus(ctx context.Context, tx pgx.Tx, urlFrontier []UrlFrontier) error {
	sql := `UPDATE url_frontier SET status = $1, updated_at = $2 WHERE id = $3`
	batch := &pgx.Batch{}

	for _, url := range urlFrontier {
		batch.Queue(sql, url.Status, carbon.Now().ToIso8601String(), url.Id)

	}
	res := tx.SendBatch(ctx, batch)

	return res.Close()
}

func UpdateUrlFrontierStatus(ctx context.Context, tx pgx.Tx, id string, status uint8) error {
	sql := `UPDATE url_frontier SET status = $1, updated_at = $2 WHERE id = $3`
	res, err := tx.Exec(ctx, sql, status, carbon.Now().ToIso8601String(), id)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()

	if rowsAffected == 0 {
		return nil
	}

	return nil
}

func GetUnScrapedUrlFrontier(ctx context.Context, tx pgx.Tx) ([]UrlFrontier, error) {
	sql := `SELECT * FROM url_frontier WHERE crawler = $1 AND status = $2`
	rows, err := tx.Query(ctx, sql, common.CRAWLER_NAME, URL_FRONTIER_STATUS_NEW)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urlFrontier []UrlFrontier
	for rows.Next() {
		var urlFrontierItem UrlFrontier
		err = rows.Scan(&urlFrontierItem.Id, &urlFrontierItem.Domain, &urlFrontierItem.Url, &urlFrontierItem.Crawler, &urlFrontierItem.Status, &urlFrontierItem.CreatedAt, &urlFrontierItem.UpdatedAt)
		if err != nil {
			return nil, err
		}
		urlFrontier = append(urlFrontier, urlFrontierItem)
	}

	return urlFrontier, nil
}
