package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

func ScrapeList(rawURL string) ([]string, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	} else if (url.Hostname() != "letterboxd.com") {
		return nil, fmt.Errorf("%s is not a letterboxd.com url", url)
	}
	filmURLList := []string{}
	c := colly.NewCollector()
	c.OnHTML("ul.poster-list", func(h *colly.HTMLElement) {
		// if !strings.Contains(h.Request.URL.Path, "film/") {
		h.ForEach("div.poster.film-poster", func(_ int, h *colly.HTMLElement) {
			filmURLList = append(filmURLList, h.Request.AbsoluteURL(h.Attr("data-target-link")))
			// c.Visit(filmPage)
		})
		// }
	})
	c.OnHTML(".next", func(h *colly.HTMLElement) {
		c.Visit(h.Request.AbsoluteURL(h.Attr("href")))
	})
	// c.OnHTML("a.micro-button.track-event", func(h *colly.HTMLElement) {
	// 	if h.Text == "TMDb" {
	// 		filmURLList = append(filmURLList, h.Attr("href"))
	// 	}
	// })
	c.Visit(url.String())
	// fmt.Println(filmURLList)
	return filmURLList, nil
}

func ScrapeFilmID(rawURL string) (int, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return -1, err
	} else if (url.Hostname() != "letterboxd.com") {
		return -1, fmt.Errorf("%s is not a letterboxd.com url", url)
	}
	c := colly.NewCollector()
	id := -1
	c.OnHTML("a.micro-button.track-event", func(h *colly.HTMLElement) {
		if h.Text == "TMDb" {
			tmdbURL, err := url.Parse(h.Attr("href"))
			if err != nil {
				panic(err)
			}
			id, err = strconv.Atoi(strings.Split(tmdbURL.Path, "/")[2])
			if err != nil {
				panic(err)
			}
		}
	})
	c.Visit(url.String())
	if id != -1 {
		return id, nil
	} else {
		return 0, fmt.Errorf("Could not find film! Perhaps %s isn't a valid film URL.", rawURL)
	}
}
