package app

import (
	"path/filepath"
	"testing"
)

func TestPosterFileName(t *testing.T) {
	tests := []struct {
		title    string
		year     uint
		expected string
	}{
		{"Star Wars", 1977, "starwars_1977.jpg"},
		{"Mad Max: Fury Road", 2015, "madmaxfuryroad_2015.jpg"},
		{"Film & Title!", 2023, "filmtitle_2023.jpg"},
		{"123 Movie", 2020, "123movie_2020.jpg"},
	}

	for _, tt := range tests {
		f := Film{Title: tt.title, Year: tt.year}
		got := posterFileName(f)
		if filepath.Base(got) != tt.expected {
			t.Errorf("posterFileName(%v) = %v, want to end with %v", tt.title, got, tt.expected)
		}
	}
}
