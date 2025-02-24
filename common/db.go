package common

import (
	"errors"
	"lexicon/indonesia-supreme-court-crawler/repository"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	Pool          *pgxpool.Pool
	StorageClient *storage.Client
	Query         *repository.Queries
)

func SetDatabase(newPool *pgxpool.Pool) error {

	if newPool == nil {
		return errors.New("cannot assign nil database")
	}
	Pool = newPool
	return nil
}

func SetStorageClient(newClient *storage.Client) error {
	if newClient == nil {
		return errors.New("cannot assign nil storage client")
	}
	StorageClient = newClient
	return nil
}

func SetQuery(newQuery *repository.Queries) error {
	if newQuery == nil {
		return errors.New("cannot assign nil query")
	}
	Query = newQuery
	return nil
}
