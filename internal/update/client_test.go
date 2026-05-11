package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLatestRelease(t *testing.T) {
	want := Release{
		TagName: "v3.0.1",
		Name:    "gflow v3.0.1",
		Body:    "- fix: something\n- feat: other",
		Assets: []Asset{
			{Name: "gflow_3.0.1_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/asset.tar.gz"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases/latest" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	origBase := apiBase
	apiBase = srv.URL
	defer func() { apiBase = origBase }()

	got, err := LatestRelease(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TagName != want.TagName {
		t.Errorf("TagName: got %q, want %q", got.TagName, want.TagName)
	}
	if len(got.Assets) != len(want.Assets) {
		t.Errorf("Assets len: got %d, want %d", len(got.Assets), len(want.Assets))
	}
}

func TestLatestRelease_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	origBase := apiBase
	apiBase = srv.URL
	defer func() { apiBase = origBase }()

	_, err := LatestRelease(context.Background(), "owner", "repo")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}
