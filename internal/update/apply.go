package update

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ApplyHomebrew delegates the update to `brew upgrade gflow`.
func ApplyHomebrew(ctx context.Context) error {
	brew, err := exec.LookPath("brew")
	if err != nil {
		return fmt.Errorf("brew not found in PATH: %w", err)
	}
	cmd := exec.CommandContext(ctx, brew, "upgrade", "gflow")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ApplyBinary downloads the release asset for the current platform, verifies its
// SHA256 checksum, extracts the binary, and atomically replaces targetPath.
func ApplyBinary(ctx context.Context, release Release, targetPath string) error {
	assetName := platformAsset(release.TagName)
	asset, ok := findAsset(release.Assets, assetName)
	if !ok {
		return fmt.Errorf("no asset %q in release %s", assetName, release.TagName)
	}

	checksumAsset, ok := findAsset(release.Assets, "checksums.txt")
	if !ok {
		return fmt.Errorf("checksums.txt not found in release %s", release.TagName)
	}

	tmpDir, err := os.MkdirTemp("", "gflow-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tarPath := filepath.Join(tmpDir, assetName)
	if err := download(ctx, asset.BrowserDownloadURL, tarPath); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	if err := verifyChecksum(ctx, checksumAsset.BrowserDownloadURL, assetName, tarPath); err != nil {
		return fmt.Errorf("checksum mismatch: %w", err)
	}

	newBin := filepath.Join(tmpDir, "gflow.new")
	if err := extractBinary(tarPath, newBin); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	if err := os.Chmod(newBin, 0o755); err != nil {
		return err
	}

	if err := atomicReplace(newBin, targetPath); err != nil {
		return fmt.Errorf("replace binary: %w — try running with sudo", err)
	}
	return nil
}

func platformAsset(version string) string {
	os_ := runtime.GOOS
	arch := runtime.GOARCH
	// goreleaser default: gflow_v3.0.0_darwin_arm64.tar.gz
	tag := strings.TrimPrefix(version, "v")
	return fmt.Sprintf("gflow_%s_%s_%s.tar.gz", tag, os_, arch)
}

func findAsset(assets []Asset, name string) (Asset, bool) {
	for _, a := range assets {
		if a.Name == name {
			return a, true
		}
	}
	return Asset{}, false
}

func download(ctx context.Context, url, dest string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func verifyChecksum(ctx context.Context, checksumURL, assetName, localPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var expected string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == assetName {
			expected = fields[0]
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("checksum for %s not found in checksums.txt", assetName)
	}

	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))

	if actual != expected {
		return fmt.Errorf("expected %s, got %s", expected, actual)
	}
	return nil
}

func extractBinary(tarPath, destPath string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		// Match the bare binary name (gflow or gflow.exe), skip path prefix.
		base := filepath.Base(hdr.Name)
		if base != "gflow" && base != "gflow.exe" {
			continue
		}
		out, err := os.Create(destPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	}
	return fmt.Errorf("gflow binary not found inside %s", tarPath)
}

func atomicReplace(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}
	// Cross-device (different FS): copy + replace.
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}
