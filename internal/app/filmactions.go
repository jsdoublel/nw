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
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/hugolgst/rich-go/client"
)

const (
	PosterPathPrefix = "https://image.tmdb.org/t/p/original/"
	DiscordRPCid     = "1223146234538360906"
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
		return fmt.Errorf("%w for film %s, status code %d != %d", ErrRetreivingPoster, fr.Title, resp.StatusCode, http.StatusOK)
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

func (app *Application) StartDiscordRPC(fr FilmRecord) error {
	ctx, cancel := context.WithCancel(context.Background())
	app.rpcCancel = cancel
	go func() {
		defer cancel()
		if err := client.Login(DiscordRPCid); err != nil {
			log.Printf("discord rpc login failed, %s", err)
			return
		}
		defer client.Logout()
		startT := time.Now()
		if err := client.SetActivity(client.Activity{
			Details:    fr.String(),
			LargeImage: PosterPathPrefix + fr.Details.PosterPath,
			LargeText:  fr.String(),
			SmallImage: "tmdb_logo",
			SmallText:  "The Movie Database",
			Timestamps: &client.Timestamps{Start: &startT},
		}); err != nil {
			log.Printf("discord rpc update failed, %s", err)
		}
		<-ctx.Done()
	}()
	return nil
}

func (app *Application) StopDiscordRPC() {
	if app.rpcCancel != nil {
		app.rpcCancel()
	}
}
