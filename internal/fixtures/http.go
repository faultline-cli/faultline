package fixtures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
)

type jsonRequestOptions struct {
	AcceptHeader        string
	OptionalStatusCodes []int
}

func getJSON(ctx context.Context, client *http.Client, rawURL string, target any, opts jsonRequestOptions) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	if strings.TrimSpace(opts.AcceptHeader) != "" {
		req.Header.Set("Accept", opts.AcceptHeader)
	}
	req.Header.Set("User-Agent", "faultline-fixtures/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		statusErr := &httpStatusError{
			URL:        rawURL,
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       strings.TrimSpace(string(body)),
		}
		for _, code := range opts.OptionalStatusCodes {
			if resp.StatusCode == code {
				return statusErr
			}
		}
		return statusErr
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func getJSONOptional(ctx context.Context, client *http.Client, rawURL string, target any, opts jsonRequestOptions) error {
	err := getJSON(ctx, client, rawURL, target, opts)
	if err == nil {
		return nil
	}
	var statusErr *httpStatusError
	if errors.As(err, &statusErr) {
		for _, code := range opts.OptionalStatusCodes {
			if statusErr.StatusCode == code {
				return nil
			}
		}
	}
	return err
}

type httpStatusError struct {
	URL        string
	StatusCode int
	Status     string
	Body       string
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("fetch %s: %s %s", path.Base(e.URL), e.Status, e.Body)
}
