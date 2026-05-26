package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type client struct {
	httpClient *http.Client
}

func NewClient() *client {
	return &client{
		httpClient: &http.Client{},
	}
}

func (c *client) Post(ctx context.Context, url string, body interface{}) (Response, error) {
	var bodyBytes []byte
	bodyString, ok := body.(string)
	if ok {
		bodyBytes = []byte(bodyString)
	} else {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return Response{}, njnerror.Wrapf("libhttp.client.Post: failed to marshal body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return Response{}, njnerror.Wrapf("libhttp.client.Post: failed to create request: %w", err)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, njnerror.Wrapf("libhttp.client.Post: failed to do request: %w", err)
	}

	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return Response{}, njnerror.Wrapf("libhttp.client.Post: failed to read response body: %w", err)
	}

	responseBodyString := string(responseBodyBytes)

	return Response{
		Code: response.StatusCode,
		Body: &responseBodyString,
	}, nil
}
