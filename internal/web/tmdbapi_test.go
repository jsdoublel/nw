package web

import (
	"testing"
)

func TestTMDBFilm(t *testing.T) {
	testCases := []struct {
		name     string
		id       int
		expected string
	}{
		{
			name:     "Dancer in the Dark",
			id:       16,
			expected: "Dancer in the Dark",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			film, err := TMDBFilm(test.id)
			if err != nil {
				t.Errorf("Produced Error %s", err)
			}
			if film.Title != test.expected {
				t.Errorf("%s != %s", film.Title, test.expected)
			}
		})
	}
}

func BenchmarkTMDBFilm(b *testing.B) {
	for b.Loop() {
		if _, err := TMDBFilm(16); err != nil {
			b.Fatalf("failed to get film data, %s", err)
		}
	}
}
