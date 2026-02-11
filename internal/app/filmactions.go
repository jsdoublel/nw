package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/jsdoublel/rich-go/client"
)

const (
	TMDBFilmPathPrefix = "https://www.themoviedb.org/movie/"
	PosterPathPrefix   = "https://image.tmdb.org/t/p/original/"

	DiscordRPCid = "1223146234538360906"
)

var (
	ErrMissingPosterPath = errors.New("poster path is missing")
	ErrRetreivingPoster  = errors.New("could not retrieve poster")
)

type DiscordRPC struct {
	name string             // string for film that is being watched
	stop context.CancelFunc // function to stop watching
}

func (d DiscordRPC) String() string {
	return d.name
}

func (d DiscordRPC) Watching() bool {
	return d.stop != nil
}

// Downloads poster given a film record containing a PosterPath in its details.
// Returns the path the poster was saved to.
func DownloadPoster(fr FilmRecord) (string, error) {
	if fr.Details.PosterPath == "" {
		return "", fmt.Errorf("%w for film %s", ErrMissingPosterPath, fr.Title)
	}
	resp, err := http.Get(PosterPathPrefix + fr.Details.PosterPath)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return "", fmt.Errorf("%w for film %s, %w", ErrRetreivingPoster, fr.Title, err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w for film %s, status code %d != %d", ErrRetreivingPoster, fr.Title, resp.StatusCode, http.StatusOK)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	path := posterFileName(fr.Film)
	return path, os.WriteFile(path, content, 0o644)
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func posterFileName(f Film) string {
	cleanTitle := nonAlphanumericRegex.ReplaceAllString(f.Title, "")
	fName := fmt.Sprintf("%s_%d.jpg", strings.ToLower(cleanTitle), f.Year)
	posterBaseDir := Config.Directories.Posters
	if posterBaseDir == "" {
		posterBaseDir = xdg.UserDirs.Download
	}
	return filepath.Join(posterBaseDir, fName)
}

func (app *Application) StartDiscordRPC(fr FilmRecord) error {
	if app.DiscordRPC.Watching() {
		app.StopDiscordRPC()
	}
	ctx, cancel := context.WithCancel(context.Background())
	app.DiscordRPC = DiscordRPC{name: fr.String(), stop: cancel}
	go func() {
		defer cancel()
		if err := client.Login(DiscordRPCid); err != nil {
			log.Printf("discord rpc login failed, %s", err)
			return
		}
		defer client.Logout()
		startT := time.Now()
		activity := client.Activity{
			Details:    fr.String(),
			State:      fr.DirectorString(),
			LargeText:  fr.String(),
			SmallImage: "tmdb_logo",
			SmallText:  "The Movie Database",
			Timestamps: &client.Timestamps{Start: &startT},
		}
		if fr.Details.PosterPath != "" {
			activity.LargeImage = PosterPathPrefix + fr.Details.PosterPath
		}
		if err := client.SetActivity(activity); err != nil {
			log.Printf("discord rpc update failed, %s", err)
		}
		<-ctx.Done()
	}()
	return nil
}

func (app *Application) StopDiscordRPC() {
	if app.DiscordRPC.stop != nil {
		app.DiscordRPC.stop()
	}
	app.DiscordRPC = DiscordRPC{}
}
