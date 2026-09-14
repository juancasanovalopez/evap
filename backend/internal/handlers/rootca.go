package handlers

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// amazonRootCA1URL is AWS IoT's publicly published trust anchor; it is a
// long-lived, non-secret file (fixed URL, no user input involved).
const amazonRootCA1URL = "https://www.amazontrust.com/repository/AmazonRootCA1.pem"

var (
	rootCAMu   sync.Mutex
	rootCABody []byte
)

// fetchRootCA returns the cached certificate, fetching and caching it on the
// first successful call; failures are not cached so a later request retries.
func fetchRootCA() ([]byte, error) {
	rootCAMu.Lock()
	defer rootCAMu.Unlock()
	if rootCABody != nil {
		return rootCABody, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(amazonRootCA1URL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status fetching root CA: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	rootCABody = body
	return rootCABody, nil
}

// RootCA serves the Amazon Root CA 1 certificate from our own domain so
// sensor owners never need to visit the AWS console or an external site to
// provision their device's trust store.
func RootCA(w http.ResponseWriter, r *http.Request) {
	body, err := fetchRootCA()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not fetch root CA certificate"})
		return
	}
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", `attachment; filename="AmazonRootCA1.pem"`)
	_, _ = w.Write(body)
}
