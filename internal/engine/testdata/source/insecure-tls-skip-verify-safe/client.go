package client

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

func NewSecureClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	return &http.Client{Transport: transport}
}

func FetchSecure(url string) (string, error) {
	client := NewSecureClient()
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	return resp.Status, nil
}
