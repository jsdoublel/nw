package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/mod/semver"
)

const latestReleaseUrl = "https://api.github.com/repos/jsdoublel/nw/releases/latest"

var releaseClient = &http.Client{Timeout: 5 * time.Second}

type Release struct {
	Version string `json:"tag_name"`
	Changes string `json:"body"`
}

// checks for new release from GitHub
func CheckSoftwareUpdate() (bool, Release) {
	latest := getLatestRelease()
	if latest.Version == "" {
		return false, Release{}
	}
	newer, err := isNewerVersion(Version, latest.Version)
	if err != nil {
		log.Printf("skipping update check, %s", err)
		return false, Release{}
	}
	if !newer {
		return false, Release{}
	}
	return true, latest
}

func getLatestRelease() Release {
	req, err := http.NewRequest(http.MethodGet, latestReleaseUrl, nil)
	if err != nil {
		log.Printf("failed to build update check request, %s", err)
		return Release{}
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := releaseClient.Do(req)
	if err != nil {
		log.Printf("failed to check for new release, %s", err)
		return Release{}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		log.Printf("unexpected status %d checking for new release", resp.StatusCode)
		return Release{}
	}
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.Printf("failed to decode release info, %s", err)
		return Release{}
	}
	return release
}

func isNewerVersion(current, latest string) (bool, error) {
	if !semver.IsValid(current) {
		return false, fmt.Errorf("current version %q is not a valid release tag", current)
	}
	if !semver.IsValid(latest) {
		return false, fmt.Errorf("latest version %q is not a valid release tag", latest)
	}
	return semver.Compare(latest, current) > 0, nil
}
