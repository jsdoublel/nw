package app

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
)

// How requently checking Letterboxd RSS is allowed
const RSSCheckTime = 15 * time.Minute

// Set of recent diary entries.
type RecentActivity struct {
	Activity  []RSSItem
	LastCheck time.Time
}

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

func (app *Application) CanCheckRSS() bool {
	return time.Since(app.RecentActivity.LastCheck) > RSSCheckTime && !Config.Features.DisableQuickUpdates
}

func (app *Application) QuickUpdateWatched() error {
	if !app.CanCheckRSS() {
		return nil
	}
	log.Printf("executing quick update with RSS...")
	app.RecentActivity.LastCheck = time.Now()
	updatedActivity, err := getRSSItems(app.Username)
	if err != nil {
		return err
	}
	return app.updateRecentActivity(updatedActivity)
}

func (app *Application) updateRecentActivity(updatedActivity []RSSItem) error {
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
	app.RecentActivity.Activity = updatedActivity
	watchedFilms, err := GetFilmsForActivities(newEntries)
	if err != nil {
		return err
	}
	for _, f := range watchedFilms { // register in film store
		if _, ok := app.Watchlist[f.LBxdID]; ok {
			app.FilmStore.deregister(f)
		}
		app.FilmStore.register(f)
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

func GetFilmsForActivities(activities []RSSItem) ([]Film, error) {
	films := make([]Film, 0, len(activities))
	for _, a := range activities {
		f, err := filmForActivity(a)
		if err != nil {
			log.Printf("failed to get Letterboxd ID for film, %s", err)
			continue
		}
		films = append(films, f)
	}
	return films, nil
}

func filmForActivity(activity RSSItem) (Film, error) {
	parsedUrl, err := url.Parse(activity.Link)
	if err != nil {
		return Film{}, fmt.Errorf("%w, link from RSS item could not be parsed as url, %s", ErrParsingRSS, activity.Link)
	}
	parts := strings.Split(strings.Trim(parsedUrl.Path, "/"), "/")
	filmUrl, err := url.JoinPath(LetterboxdUrl, "film", parts[2])
	if err != nil {
		return Film{}, fmt.Errorf("%w, couldn't create film URL from RSS link, %s", ErrParsingRSS, activity.Link)
	}
	return ScrapeFilmFromFilmPage(filmUrl)
}

func (app *Application) LastActivity() (RSSItem, error) {
	if len(app.RecentActivity.Activity) == 0 {
		return RSSItem{}, ErrNoActivity
	}
	return app.RecentActivity.Activity[0], nil
}
