package services

import (
	"context"
	"lexicon/indonesia-supreme-court-crawler/common"
	"lexicon/indonesia-supreme-court-crawler/crawler/models"
)

func UpsertUrl(urlFrontier []models.UrlFrontier) error {

	ctx := context.Background()
	tx, err := common.Pool.Begin(ctx)

	if err != nil {
		return err
	}

	err = models.UpsertUrlFrontier(ctx, tx, urlFrontier)
	if err != nil {
		return err
	}
	tx.Commit(ctx)

	return nil
}

func GetUnscrapedUrlFrontier() ([]models.UrlFrontier, error) {

	ctx := context.Background()
	tx, err := common.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	list, err := models.GetUnScrapedUrlFrontier(ctx, tx)
	if err != nil {
		return nil, err
	}

	tx.Commit(ctx)
	return list, nil

}
