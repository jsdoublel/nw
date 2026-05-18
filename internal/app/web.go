package app

import (
	"fmt"
	"io"
	"net/http"
)

func getUrlContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d != %d", resp.StatusCode, http.StatusOK)
	}
	return io.ReadAll(resp.Body)
}
