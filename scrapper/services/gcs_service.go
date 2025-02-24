package services

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"

	"lexicon/indonesia-supreme-court-crawler/common"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"cloud.google.com/go/storage"
)

func HandlePdf(ctx context.Context, name string, pdfUrl string) (lo.Tuple3[string, string, int64], error) {

	//sanitize name
	nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	fileName := nonAlphanumericRegex.ReplaceAllString(name, "_")
	size := math.Min(float64(len(fileName)), float64(100))
	sanitizedName := strings.ReplaceAll(fileName[:int(size)], " ", "_")
	pdfLocalPath, err := downloadFile(pdfUrl, sanitizedName, "pdf")
	if err != nil {
		log.Error().Err(err).Msg("Error downloading pdf")
		return lo.Tuple3[string, string, int64]{}, nil
	}
	path, err := uploadToGCS(ctx, common.StorageClient, common.GCS_BUCKET, pdfLocalPath.A, sanitizedName, common.GCS_FOLDER)
	if err != nil {
		log.Error().Err(err).Msg("Error uploading pdf to gcs")
		return lo.Tuple3[string, string, int64]{}, nil

	}

	if err := os.Remove(pdfLocalPath.A); err != nil {
		log.Error().Err(err).Msg("Error removing pdf file")
		return lo.Tuple3[string, string, int64]{}, nil

	}

	return lo.T3(name, path, pdfLocalPath.B), nil
}

func HandleHtml(ctx context.Context, name string, url string) (lo.Tuple3[string, string, int64], error) {

	//sanitize name
	nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	fileName := nonAlphanumericRegex.ReplaceAllString(name, "_")
	size := math.Min(float64(len(fileName)), float64(100))
	sanitizedName := strings.ReplaceAll(fileName[:int(size)], " ", "_")
	pdfLocalPath, err := downloadFile(url, sanitizedName, "html")
	if err != nil {
		log.Error().Err(err).Msg("Error downloading pdf")
		return lo.Tuple3[string, string, int64]{}, nil
	}
	path, err := uploadToGCS(ctx, common.StorageClient, common.GCS_BUCKET, pdfLocalPath.A, sanitizedName, common.GCS_HTML_FOLDER)
	if err != nil {
		log.Error().Err(err).Msg("Error uploading pdf to gcs")
		return lo.Tuple3[string, string, int64]{}, nil

	}

	if err := os.Remove(pdfLocalPath.A); err != nil {
		log.Error().Err(err).Msg("Error removing pdf file")
		return lo.Tuple3[string, string, int64]{}, nil

	}

	return lo.T3(name, path, pdfLocalPath.B), nil
}

func uploadToGCS(ctx context.Context, client *storage.Client, bucketName, filepath, objectName, gcsFolder string) (string, error) {
	r, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer r.Close()
	bucket := client.Bucket(bucketName)
	path := fmt.Sprintf("%s/%s", gcsFolder, objectName)
	obj := bucket.Object(path)

	wc := obj.NewWriter(ctx)
	if _, err := io.Copy(wc, r); err != nil {
		return "", err
	}

	defer wc.Close()

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, path), nil
}

func downloadFile(url string, title string, extension string) (lo.Tuple2[string, int64], error) {
	response, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Msg("Error downloading pdf: " + url)
		return lo.Tuple2[string, int64]{}, err
	}

	defer response.Body.Close()

	out, err := os.CreateTemp("", title+"."+extension)
	if err != nil {
		log.Error().Err(err).Msg("Error creating temp file")
		return lo.Tuple2[string, int64]{}, err
	}

	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error writing to temp file")
		return lo.Tuple2[string, int64]{}, err
	}

	fi, err := out.Stat()
	if err != nil {
		log.Error().Err(err).Msg("Error getting file info")
		return lo.Tuple2[string, int64]{}, err
	}

	return lo.Tuple2[string, int64]{
		A: out.Name(),
		B: fi.Size(),
	}, nil
}
