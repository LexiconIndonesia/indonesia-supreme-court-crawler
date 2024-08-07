package common

import (
	"errors"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	Pool          *pgxpool.Pool
	StorageClient *storage.Client
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
