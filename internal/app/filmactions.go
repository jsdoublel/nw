package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

const (
	PosterPathPrefix = "https://image.tmdb.org/t/p/original/"
	RpcID            = "1223146234538360906"
)

var (
	ErrMissingPosterPath = errors.New("poster path is missing")
	ErrRetreivingPoster  = errors.New("could not retrieve poster")
)

// Downloads poster given a film record containing a PosterPath in its details.
// Returns the path the poster was saved to.
func DownloadPoster(fr FilmRecord) error {
	if fr.Details.PosterPath == "" {
		return fmt.Errorf("%w for film %s", ErrMissingPosterPath, fr.Title)
	}
	resp, err := http.Get(PosterPathPrefix + fr.Details.PosterPath)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return fmt.Errorf("%w for film %s, %w", ErrRetreivingPoster, fr.Title, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s for film %s, status code %d != %d", ErrRetreivingPoster, fr.Title, resp.StatusCode, http.StatusOK)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(posterFileName(fr.Film), content, 0o644)
}

func posterFileName(f Film) string {
	fName := fmt.Sprintf("%s%d.jpg", strings.ReplaceAll(f.Title, " ", ""), f.Year)
	return filepath.Join(xdg.UserDirs.Download, fName)
}

func SetDiscordRPC(fr FilmRecord) error {
	return nil
}
