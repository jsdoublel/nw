package filmdata

import (
	"reflect"
	"testing"
)

func TestScrapeList(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected []string
	}{
		{
			name: "oscars: 2024",
			url:  "https://letterboxd.com/oscars/list/the-96th-academy-award-nominees-for-best/",
			expected: []string{
				"https://letterboxd.com/film/oppenheimer-2023/",
				"https://letterboxd.com/film/american-fiction/",
				"https://letterboxd.com/film/anatomy-of-a-fall/",
				"https://letterboxd.com/film/barbie/",
				"https://letterboxd.com/film/the-holdovers/",
				"https://letterboxd.com/film/killers-of-the-flower-moon/",
				"https://letterboxd.com/film/maestro-2023/",
				"https://letterboxd.com/film/past-lives/",
				"https://letterboxd.com/film/poor-things-2023/",
				"https://letterboxd.com/film/the-zone-of-interest/",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			urlList, err := ScrapeList(test.url)
			if err != nil {
				t.Errorf("Produced Error %s", err)
			}
			if !reflect.DeepEqual(test.expected, urlList) {
				t.Errorf("want=%v\n!= got=\n%v", test.expected, urlList)
			}
		})
	}
}

func BenchmarkScrapeList(b *testing.B) {
	testListUrl := "https://letterboxd.com/sentralperk/list/sight-sound/"
	for b.Loop() {
		if _, err := ScrapeList(testListUrl); err != nil {
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
