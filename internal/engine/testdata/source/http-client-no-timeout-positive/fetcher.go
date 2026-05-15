package fetcher

import (
	"fmt"
	"io"
	"net/http"
)

const apiURL = "https://api.example.com/data"

func FetchData(path string) ([]byte, error) {
	resp, err := http.Get(apiURL + path)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
