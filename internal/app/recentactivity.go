package app

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Set of recent diary entries.
type RecentActivity []RSSItem

type RSSItem struct {
	Guid string `xml:"guid"`
	Link string `xml:"link"`
}

var (
	ErrRetreivingRSS = errors.New("could not retrieve RSS")
	ErrParsingRSS    = errors.New("failed to parse RSS")
	ErrNoActivity    = errors.New("no recent activity")
)

func getRSSItems(username string) ([]RSSItem, error) {
	rssUrl, err := url.JoinPath(LetterboxdUrl, username, "rss")
	if err != nil {
		return nil, fmt.Errorf("%w, problem joining url parts, %w", ErrRetreivingRSS, err)
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

func (app *Application) QuickUpdateWatched() error {
	updatedActivity, err := getRSSItems(app.Username)
	if err != nil {
		return err
	}
	newEntries := make([]RSSItem, 0)
	for _, item := range updatedActivity {
		if lastActivity, err := app.LastActivity(); err == nil && lastActivity.Guid == item.Guid {
			break
		}
		if !strings.HasPrefix(item.Guid, "letterboxd-review") && !strings.HasPrefix(item.Guid, "letterboxd-watch") {
			break
		}
		newEntries = append(newEntries, item)
	}
	if len(newEntries) == 0 {
		return nil
	}
	app.RecentActivity = updatedActivity
	watchedFilms, err := GetFilmsForActivities(newEntries)
	if err != nil {
		return err
	}
	app.Watchlist.RemoveFilms(watchedFilms)
	app.WatchedFilms.AddFilms(watchedFilms)
	if err := app.updateNextWatchQueue(); err != nil {
		return err
	}
	if err := app.updateTrackedLists(false); err != nil {
		return err
	}
	return nil
}

func GetFilmsForActivities(activity []RSSItem) ([]Film, error) {
	return nil, nil
}

func (app *Application) LastActivity() (RSSItem, error) {
	if app.RecentActivity == nil || len(app.RecentActivity) > 0 {
		return RSSItem{}, ErrNoActivity
	}
	return app.RecentActivity[0], nil
}
