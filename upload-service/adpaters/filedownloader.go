package adapters

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type fileDownloader struct {
	client *http.Client
}

func NewFileDownloader(client *http.Client) *fileDownloader {
	return &fileDownloader{
		client: client,
	}
}

func (fileDownloader *fileDownloader) DownloadFile(ctx context.Context, url string) (io.ReadCloser, int64, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("filedownloader.DownloadFile: failed to create request: %w", err)
	}

	response, err := fileDownloader.client.Do(request)
	if err != nil {
		return nil, 0, fmt.Errorf("filedownloader.DownloadFile: failed to download file: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("filedownloader.DownloadFile: failed to download file: %s", response.Status)
	}

	return response.Body, response.ContentLength, nil
}
