package services

import (
	"context"
	"io"
	"lexicon/indonesia-supreme-court-crawler/common"
	"lexicon/indonesia-supreme-court-crawler/scrapper/models"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"

	"cloud.google.com/go/storage"
)

func HandlePdf(metadata models.Metadata, pdfUrl, pdfName string) (string, error) {
	pdfPath, err := downloadPdf(pdfUrl)
	if err != nil {
		log.Error().Err(err).Msg("Error downloading pdf")
	}
	ctx := context.Background()
	path, err := uploadToGCS(ctx, common.StorageClient, common.GCS_BUCKET, pdfPath, pdfName)
	if err != nil {
		log.Error().Err(err).Msg("Error uploading pdf to gcs")
	}

	return path, nil
}

func UpsertExtraction(extraction models.Extraction) error {
	ctx := context.Background()
	tx, err := common.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	err = models.UpsertExtraction(ctx, tx, extraction)
	if err != nil {
		return err
	}

	tx.Commit(ctx)

	return nil
}

func uploadToGCS(ctx context.Context, client *storage.Client, bucketName, filepath, objectName string) (string, error) {

	r, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer r.Close()
	wc := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(wc, r); err != nil {
		return "", err
	}

	defer wc.Close()

	return wc.MediaLink, nil
}

func downloadPdf(url string) (string, error) {

	response, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Msg("Error downloading pdf")
		return "", err
	}

	defer response.Body.Close()

	// Create Temp File
	out, err := os.CreateTemp("", "pdf")
	if err != nil {
		log.Error().Err(err).Msg("Error creating temp file")
		return "", err
	}

	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error writing to temp file")
		return "", err
	}

	return out.Name(), nil
}
