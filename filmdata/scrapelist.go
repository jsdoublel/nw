package filmdata

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

const LetterboxdUrl = "https://letterboxd.com"

type FilmListHeader struct {
	Name    string
	ListUrl string
}

func (flh FilmListHeader) String() string {
	return fmt.Sprintf("{%s : %s}", flh.Name, flh.ListUrl)
}

func MakeUserWatchlistHeader(username string) (FilmListHeader, error) {
	watchlistUrl, err := url.JoinPath(LetterboxdUrl, username, "watchlist")
	if err != nil {
		return FilmListHeader{}, fmt.Errorf("problem joining url parts, %w", err)
	}
	return FilmListHeader{Name: "Watchlist", ListUrl: watchlistUrl}, nil
}

func MakeUserFilmListHeader(username string) (FilmListHeader, error) {
	filmListUrl, err := url.JoinPath(LetterboxdUrl, username, "films")
	if err != nil {
		return FilmListHeader{}, fmt.Errorf("problem joining url parts, %w", err)
	}
	return FilmListHeader{Name: "Watched", ListUrl: filmListUrl}, nil
}

func ScapeUserLists(username string) ([]FilmListHeader, error) {
	listPageUrl, err := url.JoinPath(LetterboxdUrl, username, "lists")
	if err != nil {
		return nil, fmt.Errorf("problem joining url parts, %w", err)
	}
	usersListUrls := []FilmListHeader{}
	c := colly.NewCollector()
	c.OnHTML("h2.name.prettify", func(h *colly.HTMLElement) {
		h.ForEach("a[href]", func(_ int, h *colly.HTMLElement) {
			if listUrl := h.Request.AbsoluteURL(h.Attr("href")); strings.Contains(listUrl, "/list/") {
				usersListUrls = append(usersListUrls, FilmListHeader{Name: h.Text, ListUrl: listUrl})
			}
		})
	})
	var paginationErr error
	c.OnHTML("a.next", func(h *colly.HTMLElement) {
		if paginationErr != nil {
			return
		}
		nextURL := h.Request.AbsoluteURL(h.Attr("href"))
		if err := c.Visit(nextURL); err != nil && !errors.Is(err, colly.ErrAlreadyVisited) {
			paginationErr = fmt.Errorf("paginate user lists: %w", err)
		}
	})
	if err := c.Visit(listPageUrl); err != nil {
		return nil, fmt.Errorf("problem trying to visit url, %w", err)
	}
	if paginationErr != nil {
		return nil, paginationErr
	}
	return usersListUrls, nil
}

func ScrapeList(rawURL string) ([]string, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	} else if url.Hostname() != "letterboxd.com" {
		return nil, fmt.Errorf("%s is not a letterboxd.com url", url)
	}
	filmURLList := []string{}
	c := colly.NewCollector()
	c.OnHTML("ul.poster-list", func(h *colly.HTMLElement) {
		h.ForEach("div.react-component", func(_ int, h *colly.HTMLElement) {
			if fUrl := h.Request.AbsoluteURL(h.Attr("data-target-link")); strings.Contains(fUrl, "/film/") {
				filmURLList = append(filmURLList, fUrl)
			}
		})
	})
	var paginationErr error
	c.OnHTML(".next", func(h *colly.HTMLElement) {
		if paginationErr != nil {
			return
		}
		nextURL := h.Request.AbsoluteURL(h.Attr("href"))
		if err := c.Visit(nextURL); err != nil && !errors.Is(err, colly.ErrAlreadyVisited) {
			paginationErr = fmt.Errorf("paginate list: %w", err)
		}
	})
	if err = c.Visit(url.String()); err != nil {
		return nil, err
	}
	if paginationErr != nil {
		return nil, paginationErr
	}
	return filmURLList, nil
}

func ScrapeFilmID(rawURL string) (int, error) {
	filmUrl, err := url.Parse(rawURL)
	if err != nil {
		return -1, err
	} else if filmUrl.Hostname() != "letterboxd.com" {
		return -1, fmt.Errorf("%s is not a letterboxd.com url", filmUrl)
	}
	c := colly.NewCollector()
	id := 0
	c.OnHTML("a.micro-button.track-event", func(h *colly.HTMLElement) {
		if h.Text == "TMDB" {
			tmdbURL, err := filmUrl.Parse(h.Attr("href"))
			if err != nil {
				panic(err)
			}
			id, err = strconv.Atoi(strings.Split(tmdbURL.Path, "/")[2])
			if err != nil {
				panic(err)
			}
		}
	})
	if err = c.Visit(filmUrl.String()); err != nil {
		return 0, err
	}
	if id != 0 {
		return id, nil
	} else {
		return 0, fmt.Errorf("could not find film! perhaps %s isn't a valid film URL", rawURL)
	}
}
