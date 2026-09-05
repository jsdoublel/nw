package app

import (
	"fmt"
	"io"
	"net/http"
)

var webHTTPClient = &http.Client{Transport: scrapeTransport}

func getUrlContent(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://letterboxd.com/")
	for k, v := range httpClient.Headers {
		if len(v) == 0 {
			continue
		}
		req.Header.Set(k, v[0])
	}
	resp, err := webHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d != %d", resp.StatusCode, http.StatusOK)
	}
	return io.ReadAll(resp.Body)
}
