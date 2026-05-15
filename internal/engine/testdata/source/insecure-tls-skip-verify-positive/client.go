package insecuretlsskipverifypositive
package client

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

func NewInsecureClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &http.Client{Transport: transport}
}

func FetchIgnoringTLS(url string) (string, error) {
	client := NewInsecureClient()
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	return resp.Status, nil
}
