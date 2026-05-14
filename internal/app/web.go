package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrRetrievingContent = errors.New("error retrieving content")

func getUrlContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, fmt.Errorf("%w for user %s, %w", ErrRetrievingContent, url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"%w for user %s, status code %d != %d",
			ErrRetrievingContent,
			url,
			resp.StatusCode,
			http.StatusOK,
		)
	}
	return io.ReadAll(resp.Body)
}
