package app

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
)

// Set of recent diary entries. Represented as map from guid number to film.
type RecentActivity map[int]Film

type RSSItem struct {
	Guid string `xml:"guid"`
	Link string `xml:"link"`
}

var (
	ErrRetreivingRSS = errors.New("could not retrieve RSS")
	ErrParsingRSS    = errors.New("failed to parse RSS")
)

func getRSSItems(username string) ([]RSSItem, error) {
	rssUrl, err := url.JoinPath(LetterboxdUrl, username, "rss")
	if err != nil {
		return nil, fmt.Errorf("problem joining url parts, %w", err)
	}
	content, err := getUrlContent(rssUrl)
	if err != nil {
		return nil, fmt.Errorf("%w for user %s, %w", ErrRetreivingRSS, username, err)
	}
	var rssItems struct {
		Items []RSSItem `xml:"channel>item"`
	}
	if err := xml.Unmarshal(content, &rssItems); err != nil {
		return nil, fmt.Errorf("%w, %s", ErrParsingRSS, err)
	}
	return rssItems.Items, nil
}
