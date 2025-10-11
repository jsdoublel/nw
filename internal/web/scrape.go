package web

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly"

	m "github.com/jsdoublel/nw/internal/model"
)

const LetterboxdUrl = "https://letterboxd.com"

var ErrBadScrape error = errors.New("bad scrape")
var ErrInvalidUrl error = errors.New("invalid url")

var titleYearRegex = regexp.MustCompile(`^(.+?)\s+\((\d{4})\)$`)

func ScapeUserLists(username string) ([]*m.FilmList, error) {
	listPageUrl, err := url.JoinPath(LetterboxdUrl, username, "lists")
	if err != nil {
		return nil, fmt.Errorf("problem joining url parts, %w", err)
	}
	usersListUrls := []*m.FilmList{}
	c := colly.NewCollector()
	c.OnHTML("h2.name.prettify", func(h *colly.HTMLElement) {
		h.ForEach("a[href]", func(_ int, h *colly.HTMLElement) {
			if listUrl := h.Request.AbsoluteURL(h.Attr("href")); strings.Contains(listUrl, "/list/") {
				usersListUrls = append(usersListUrls, &m.FilmList{Name: h.Text, ListUrl: listUrl})
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

// Scraps list name, and film urls from list url.
//
// The list name may be empty if it is not listed on the webpage (e.g., a watchlist).
func ScrapeFilmList(rawURL string) (fl m.FilmList, err error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return
	} else if url.Hostname() != "letterboxd.com" {
		err = fmt.Errorf("%s is not a letterboxd.com url", url)
		return
	}
	fl.ListUrl = rawURL
	c := colly.NewCollector()
	c.OnHTML("h1.title-1.prettify", func(h *colly.HTMLElement) {
		fl.Name = h.Text
	})
	posterScrapper := func(h *colly.HTMLElement) {
		h.ForEach("div.react-component", func(_ int, h *colly.HTMLElement) {
			if fUrl := h.Request.AbsoluteURL(h.Attr("data-target-link")); strings.Contains(fUrl, "/film/") {
				f := m.Film{Url: fUrl}
				title := h.Attr("data-item-name")
				if title != "" {
					if matches := titleYearRegex.FindStringSubmatch(title); len(matches) == 3 {
						f.Name = matches[1]
						if year, err := strconv.Atoi(matches[2]); err == nil {
							f.Year = uint(year)
						}
					}
				}
				if f.Name != "" && f.Year != 0 {
					fl.Films = append(fl.Films, &f)
				} else {
					log.Printf("failed to parse film title %s", title)
				}
			}
		})
	}
	c.OnHTML("ul.poster-list", posterScrapper)
	c.OnHTML("ul.poster-grid", posterScrapper)
	c.OnHTML("div.poster-grid", posterScrapper)
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
		return
	}
	if paginationErr != nil {
		err = paginationErr
		return
	}
	return
}

func ScrapeFilmID(rawURL string) (id int, err error) {
	filmUrl, err := url.Parse(rawURL)
	if err != nil {
		return -1, err
	} else if filmUrl.Hostname() != "letterboxd.com" {
		return -1, fmt.Errorf("%w, %s is not a letterboxd.com url", ErrInvalidUrl, filmUrl)
	}
	c := colly.NewCollector()
	var scrapingErr error
	c.OnHTML("a.micro-button.track-event", func(h *colly.HTMLElement) {
		if h.Text == "TMDB" {
			tmdbURL, err := filmUrl.Parse(h.Attr("href"))
			if err != nil {
				scrapingErr = err
			}
			id, err = strconv.Atoi(strings.Split(tmdbURL.Path, "/")[2])
			if err != nil {
				scrapingErr = err
			}
		}
	})
	if err = c.Visit(filmUrl.String()); err != nil {
		return
	}
	if scrapingErr != nil {
		err = fmt.Errorf("%w, %w", ErrBadScrape, scrapingErr)
		return
	}
	if id == 0 {
		err = fmt.Errorf("%w, did not find TMDB id when scraping %s", ErrInvalidUrl, rawURL)
	}
	return
}
