package filmdata

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

func ScrapeList(rawURL string) ([]string, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	} else if url.Hostname() != "letterboxd.com" {
		return nil, fmt.Errorf("%s is not a letterboxd.com url", url)
	}
	filmURLList := []string{}
	count := 0
	c := colly.NewCollector()
	c.OnHTML("ul.poster-list", func(h *colly.HTMLElement) {
		// if !strings.Contains(h.Request.URL.Path, "film/") {
		// h.ForEach("div.poster.film-poster", func(_ int, h *colly.HTMLElement) {
		h.ForEach("div.react-component", func(_ int, h *colly.HTMLElement) {
			count++
			if fUrl := h.Request.AbsoluteURL(h.Attr("data-target-link")); strings.Contains(fUrl, "/film/") {
				filmURLList = append(filmURLList, fUrl)
			}
			// c.Visit(filmPage)
		})
		// }
	})
	c.OnHTML(".next", func(h *colly.HTMLElement) {
		err = c.Visit(h.Request.AbsoluteURL(h.Attr("href")))
	})
	if err != nil {
		return nil, err
	}
	// c.OnHTML("a.micro-button.track-event", func(h *colly.HTMLElement) {
	// 	if h.Text == "TMDb" {
	// 		filmURLList = append(filmURLList, h.Attr("href"))
	// 	}
	// })
	if err = c.Visit(url.String()); err != nil {
		return nil, err
	}
	// fmt.Println(filmURLList)
	log.Printf("%d div.react-component found, %d were films", count, len(filmURLList))
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
