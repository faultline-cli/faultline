package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func FetchData(url string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
