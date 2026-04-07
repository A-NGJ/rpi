package upgrade

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// executablePath is the function used to resolve the current binary's path.
// Tests override this to point at a temp file.
var executablePath = os.Executable

// Asset represents a single release asset from the GitHub API.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Release represents the relevant fields from the GitHub releases/latest API response.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// FetchLatestRelease fetches the latest release metadata from the GitHub API.
// baseURL allows tests to inject an httptest server URL; production callers
// pass "https://api.github.com".
func FetchLatestRelease(baseURL, repo string) (Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", baseURL, repo)
	resp, err := http.Get(url)
	if err != nil {
		return Release{}, fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return Release{}, fmt.Errorf("decoding release JSON: %w", err)
	}
	return rel, nil
}

// CompareVersions returns true if latest is newer than current.
// The special version "dev" always returns true (needs upgrade).
// Both versions may optionally have a "v" prefix, which is stripped.
func CompareVersions(current, latest string) (bool, error) {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	if current == "dev" {
		return true, nil
	}

	curParts, err := parseVersion(current)
	if err != nil {
		return false, fmt.Errorf("parsing current version %q: %w", current, err)
	}
	latParts, err := parseVersion(latest)
	if err != nil {
		return false, fmt.Errorf("parsing latest version %q: %w", latest, err)
	}

	for i := 0; i < 3; i++ {
		if latParts[i] > curParts[i] {
			return true, nil
		}
		if latParts[i] < curParts[i] {
			return false, nil
		}
	}
	return false, nil // equal
}

// parseVersion splits a "major.minor.patch" string into three ints.
func parseVersion(v string) ([3]int, error) {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("expected 3 parts, got %d in %q", len(parts), v)
	}
	var result [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return [3]int{}, fmt.Errorf("non-numeric version component %q: %w", p, err)
		}
		result[i] = n
	}
	return result, nil
}

// DownloadAndVerify downloads the archive and checksums file, verifies the
// archive's SHA256 against the expected checksum, then extracts and returns
// the "rpi" binary from the tarball.
func DownloadAndVerify(archiveURL, checksumsURL, expectedArchiveName string) ([]byte, error) {
	// Download the archive.
	archiveData, err := httpGet(archiveURL)
	if err != nil {
		return nil, fmt.Errorf("downloading archive: %w", err)
	}

	// Download the checksums file.
	checksumsData, err := httpGet(checksumsURL)
	if err != nil {
		return nil, fmt.Errorf("downloading checksums: %w", err)
	}

	// Compute SHA256 of the archive.
	actualHash := sha256.Sum256(archiveData)
	actualHex := fmt.Sprintf("%x", actualHash)

	// Find the expected hash in the checksums file.
	expectedHex, err := findChecksum(string(checksumsData), expectedArchiveName)
	if err != nil {
		return nil, err
	}

	if actualHex != expectedHex {
		return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHex, actualHex)
	}

	// Extract the "rpi" binary from the tarball.
	binary, err := extractBinaryFromTarGz(archiveData)
	if err != nil {
		return nil, fmt.Errorf("extracting binary: %w", err)
	}
	return binary, nil
}

// findChecksum looks for a line matching the given filename in sha256sum-formatted text.
// Format: "{hash}  {filename}" (two spaces between hash and filename).
func findChecksum(checksums, filename string) (string, error) {
	for _, line := range strings.Split(checksums, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: hash  filename (two spaces)
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) == 2 && parts[1] == filename {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("checksum for %q not found in checksums file", filename)
}

// extractBinaryFromTarGz extracts the file named "rpi" from a tar.gz archive.
func extractBinaryFromTarGz(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("opening gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar entry: %w", err)
		}
		// Match "rpi" regardless of directory prefix in the archive.
		name := filepath.Base(hdr.Name)
		if name == "rpi" && hdr.Typeflag == tar.TypeReg {
			content, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("reading binary from archive: %w", err)
			}
			return content, nil
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", "rpi")
}

// ReplaceBinary atomically replaces the binary at execPath with newBinary.
// It writes to a temp file in the same directory (same filesystem) then renames.
func ReplaceBinary(execPath string, newBinary []byte) error {
	dir := filepath.Dir(execPath)
	tmp, err := os.CreateTemp(dir, "rpi-upgrade-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(newBinary); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing binary: %w", err)
	}

	return nil
}

// Upgrade orchestrates the full self-upgrade flow:
// fetch latest release, compare versions, download+verify, replace binary.
func Upgrade(w io.Writer, baseURL, repo, currentVersion string) error {
	rel, err := FetchLatestRelease(baseURL, repo)
	if err != nil {
		return err
	}

	needsUpgrade, err := CompareVersions(currentVersion, rel.TagName)
	if err != nil {
		return err
	}

	latestVersion := strings.TrimPrefix(rel.TagName, "v")
	if !needsUpgrade {
		fmt.Fprintf(w, "Already up to date (v%s)\n", latestVersion)
		return nil
	}

	// Determine the expected archive name for this platform.
	archiveName := fmt.Sprintf("rpi_%s_%s_%s.tar.gz", latestVersion, runtime.GOOS, runtime.GOARCH)

	// Find the archive and checksums asset URLs.
	var archiveURL, checksumsURL string
	for _, a := range rel.Assets {
		switch a.Name {
		case archiveName:
			archiveURL = a.BrowserDownloadURL
		case "checksums.txt":
			checksumsURL = a.BrowserDownloadURL
		}
	}

	if archiveURL == "" {
		return fmt.Errorf("no release asset found for %s", archiveName)
	}
	if checksumsURL == "" {
		return fmt.Errorf("no checksums.txt found in release assets")
	}

	fmt.Fprintf(w, "Downloading v%s...\n", latestVersion)

	binary, err := DownloadAndVerify(archiveURL, checksumsURL, archiveName)
	if err != nil {
		return err
	}

	execPath, err := executablePath()
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	if err := ReplaceBinary(execPath, binary); err != nil {
		return fmt.Errorf("%w\n\nTry: sudo rpi upgrade, or move the binary to a user-writable location", err)
	}

	currentDisplay := strings.TrimPrefix(currentVersion, "v")
	if currentDisplay == "dev" {
		currentDisplay = "dev"
	} else {
		currentDisplay = "v" + currentDisplay
	}
	fmt.Fprintf(w, "Upgraded from %s to v%s\n", currentDisplay, latestVersion)
	return nil
}

// goos returns runtime.GOOS. Exposed as a function so tests can reference it.
func goos() string {
	return runtime.GOOS
}

// goarch returns runtime.GOARCH. Exposed as a function so tests can reference it.
func goarch() string {
	return runtime.GOARCH
}

// httpGet performs an HTTP GET and returns the response body bytes.
func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}
