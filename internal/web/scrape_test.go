package web

import (
	"reflect"
	"testing"

	m "github.com/jsdoublel/nw/internal/model"
)

func TestScrapeList(t *testing.T) {
	testCases := []struct {
		name     string
		expected m.FilmList
	}{
		{
			name: "oscars: 2024",
			expected: m.FilmList{
				Name:    "The 96th Academy Award nominees for Best Motion Picture of the Year",
				ListUrl: "https://letterboxd.com/oscars/list/the-96th-academy-award-nominees-for-best/",
				Films: []*m.Film{
					{Url: "https://letterboxd.com/film/oppenheimer-2023/", Name: "Oppenheimer", Year: 2023},
					{Url: "https://letterboxd.com/film/american-fiction/", Name: "American Fiction", Year: 2023},
					{Url: "https://letterboxd.com/film/anatomy-of-a-fall/", Name: "Anatomy of a Fall", Year: 2023},
					{Url: "https://letterboxd.com/film/barbie/", Name: "Barbie", Year: 2023},
					{Url: "https://letterboxd.com/film/the-holdovers/", Name: "The Holdovers", Year: 2023},
					{Url: "https://letterboxd.com/film/killers-of-the-flower-moon/", Name: "Killers of the Flower Moon", Year: 2023},
					{Url: "https://letterboxd.com/film/maestro-2023/", Name: "Maestro", Year: 2023},
					{Url: "https://letterboxd.com/film/past-lives/", Name: "Past Lives", Year: 2023},
					{Url: "https://letterboxd.com/film/poor-things-2023/", Name: "Poor Things", Year: 2023},
					{Url: "https://letterboxd.com/film/the-zone-of-interest/", Name: "The Zone of Interest", Year: 2023},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			fl, err := ScrapeFilmList(test.expected.ListUrl)
			if err != nil {
				t.Errorf("Produced Error %s", err)
			}
			if !reflect.DeepEqual(test.expected, fl) {
				t.Errorf("want=%v\n!= got=%v\n", test.expected, fl)
			}
		})
	}
}

func BenchmarkScrapeList(b *testing.B) {
	testListUrl := "https://letterboxd.com/sentralperk/list/sight-sound/"
	for b.Loop() {
		if _, err := ScrapeFilmList(testListUrl); err != nil {
			b.Fatalf("failed to scrape list, %s", err)
		}
	}
}

func BenchmarkScrapeUserLists(b *testing.B) {
	for b.Loop() {
		if _, err := ScapeUserLists("jsdoublel"); err != nil {
			b.Fatalf("failed to scrape user's lists, %s", err)
		}
	}
}

func TestScrapeFilmID(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected int
	}{
		{
			name:     "Dancer in the Dark",
			url:      "https://letterboxd.com/film/dancer-in-the-dark/",
			expected: 16,
		},
		{
			name:     "2001: A Space Odyessey",
			url:      "https://letterboxd.com/film/2001-a-space-odyssey/",
			expected: 62,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			id, err := ScrapeFilmID(test.url)
			if err != nil {
				t.Errorf("Produced Error %s", err)
			}
			if test.expected != id {
				t.Errorf("%d != %d", test.expected, id)
			}
		})
	}
}

func BenchmarkScrapeFilmID(b *testing.B) {
	filmUrl := "https://letterboxd.com/film/2001-a-space-odyssey/"
	for b.Loop() {
		if _, err := ScrapeFilmID(filmUrl); err != nil {
			b.Fatalf("failed to scrape film url, %s", err)
		}
	}
}
