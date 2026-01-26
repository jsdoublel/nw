package app

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

const LetterboxdUrl = "https://letterboxd.com"

var (
	ErrBadScrape  error = errors.New("bad scrape")
	ErrInvalidUrl error = errors.New("invalid url")

	titleYearRegex = regexp.MustCompile(`^(.+?)\s+\((\d{4})\)$`)
)

func ScapeUserLists(username string) ([]*FilmList, error) {
	listPageUrl, err := url.JoinPath(LetterboxdUrl, username, "lists")
	if err != nil {
		return nil, fmt.Errorf("problem joining url parts, %w", err)
	}
	usersListUrls := []*FilmList{}
	c := colly.NewCollector()
	attachScrapeLogger(c, "user lists")
	c.OnHTML("div.body", func(h *colly.HTMLElement) {
		fl := &FilmList{}
		h.ForEach("h2.name.prettify a[href]", func(_ int, link *colly.HTMLElement) {
			if listUrl := link.Request.AbsoluteURL(link.Attr("href")); strings.Contains(listUrl, "/list/") {
				fl.Name = strings.TrimSpace(link.Text)
				fl.Url = listUrl
			}
		})
		h.ForEach("span.value", func(i int, h *colly.HTMLElement) {
			parts := strings.Split(h.Text, "\u00A0")
			num, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				log.Printf("couldn't parse number of films in list %s from %s", fl.Name, parts[0])
			}
			fl.NumFilms = num
		})
		fl.Desc = parseDescription(h, "p")
		if fl.Name != "" && fl.Url != "" {
			usersListUrls = append(usersListUrls, fl)
		}
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
		return nil, fmt.Errorf("problem trying to visit url %s, %w", listPageUrl, err)
	}
	if paginationErr != nil {
		return nil, paginationErr
	}
	return usersListUrls, nil
}

// Scraps list name, and film urls from list url.
//
// The list name may be empty if it is not listed on the webpage (e.g., a watchlist).
func ScrapeFilmList(rawURL string) (fl FilmList, err error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		err = fmt.Errorf("%w, %w", ErrInvalidUrl, err)
		return
	} else if url.Hostname() != "letterboxd.com" {
		err = fmt.Errorf("%w, %s is not a letterboxd.com url", ErrInvalidUrl, url)
		return
	}
	fl.Url = rawURL
	c := colly.NewCollector()
	attachScrapeLogger(c, rawURL)
	c.OnHTML("h1.title-1.prettify", func(h *colly.HTMLElement) {
		fl.Name = strings.TrimSpace(h.Text)
	})
	c.OnHTML("div.body-text", func(h *colly.HTMLElement) {
		if h.Attr("data-full-text-url") != "#list-notes" {
			return
		}
		fl.Desc = parseDescription(h, "p")
	})
	posterScrapper := func(h *colly.HTMLElement) {
		h.ForEach("div.react-component", func(_ int, h *colly.HTMLElement) {
			if fUrl := h.Request.AbsoluteURL(h.Attr("data-target-link")); strings.Contains(fUrl, "/film/") {
				f := Film{Url: fUrl}
				title := h.Attr("data-item-name")
				if matches := titleYearRegex.FindStringSubmatch(title); len(matches) == 3 {
					f.Title = strings.TrimSpace(matches[1])
					if year, err := strconv.Atoi(matches[2]); err == nil {
						f.Year = uint(year)
					}
				}
				if id, err := strconv.Atoi(h.Attr("data-film-id")); err == nil {
					f.LBxdID = id
				}
				if f.Title != "" && f.Year != 0 && f.LBxdID != 0 {
					fl.Films = append(fl.Films, &f)
				} else {
					log.Printf("failed to parse film %s from %s", fUrl, fl.Url)
				}
			}
		})
		h.ForEachWithBreak("li.posteritem.numbered-list-item", func(i int, h *colly.HTMLElement) bool {
			fl.Ordered = true
			return false
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
	fl.NumFilms = len(fl.Films)
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
	attachScrapeLogger(c, rawURL)
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

func parseDescription(h *colly.HTMLElement, selector string) string {
	var builder strings.Builder
	first := true
	h.ForEach(selector, func(_ int, node *colly.HTMLElement) {
		text := strings.TrimSpace(node.Text)
		if text == "" {
			return
		}
		if !first {
			builder.WriteString("\n\n")
		} else {
			first = false
		}
		builder.WriteString(text)
	})
	return builder.String()
}

func attachScrapeLogger(c *colly.Collector, label string) {
	c.OnError(func(resp *colly.Response, err error) {
		status := 0
		url := label
		if resp != nil {
			status = resp.StatusCode
			if resp.Request != nil && resp.Request.URL != nil {
				url = resp.Request.URL.String()
			}
		}
		log.Printf("scrape error [%s] status=%d url=%s err=%v", label, status, url, err)
	})
	c.OnResponse(func(resp *colly.Response) {
		if resp.StatusCode >= 400 {
			log.Printf("scrape response [%s] status=%d url=%s", label, resp.StatusCode, resp.Request.URL)
		}
	})
}
