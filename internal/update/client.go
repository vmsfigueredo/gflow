package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var apiBase = "https://api.github.com"

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

func LatestRelease(ctx context.Context, owner, repo string) (Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", apiBase, owner, repo)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("github api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Release{}, fmt.Errorf("no releases found for %s/%s", owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var r Release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Release{}, fmt.Errorf("decode release: %w", err)
	}
	return r, nil
}
