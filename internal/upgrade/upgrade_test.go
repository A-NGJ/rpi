package upgrade

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
		wantErr bool
	}{
		{
			name:    "newer available",
			current: "1.0.0",
			latest:  "1.1.0",
			want:    true,
		},
		{
			name:    "already current",
			current: "1.1.0",
			latest:  "1.1.0",
			want:    false,
		},
		{
			name:    "dev build always upgrades",
			current: "dev",
			latest:  "1.0.0",
			want:    true,
		},
		{
			name:    "patch upgrade",
			current: "1.0.0",
			latest:  "1.0.1",
			want:    true,
		},
		{
			name:    "major upgrade",
			current: "1.0.0",
			latest:  "2.0.0",
			want:    true,
		},
		{
			name:    "current newer than latest",
			current: "1.1.0",
			latest:  "1.0.0",
			want:    false,
		},
		{
			name:    "v prefix stripped from current",
			current: "v1.0.0",
			latest:  "1.1.0",
			want:    true,
		},
		{
			name:    "v prefix stripped from latest",
			current: "1.0.0",
			latest:  "v1.1.0",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareVersions(tt.current, tt.latest)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CompareVersions(%q, %q) error = %v, wantErr %v", tt.current, tt.latest, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestFetchLatestRelease(t *testing.T) {
	release := map[string]any{
		"tag_name": "v1.2.3",
		"assets": []map[string]any{
			{
				"name":                 "rpi_1.2.3_darwin_arm64.tar.gz",
				"browser_download_url": "https://example.com/rpi_1.2.3_darwin_arm64.tar.gz",
			},
			{
				"name":                 "checksums.txt",
				"browser_download_url": "https://example.com/checksums.txt",
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/A-NGJ/rpi/releases/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer srv.Close()

	rel, err := FetchLatestRelease(srv.URL, "A-NGJ/rpi")
	if err != nil {
		t.Fatalf("FetchLatestRelease() error: %v", err)
	}

	if rel.TagName != "v1.2.3" {
		t.Errorf("TagName = %q, want %q", rel.TagName, "v1.2.3")
	}

	if len(rel.Assets) != 2 {
		t.Fatalf("len(Assets) = %d, want 2", len(rel.Assets))
	}

	if rel.Assets[0].Name != "rpi_1.2.3_darwin_arm64.tar.gz" {
		t.Errorf("Assets[0].Name = %q, want %q", rel.Assets[0].Name, "rpi_1.2.3_darwin_arm64.tar.gz")
	}
}

// makeTarGz creates a tar.gz archive containing a single file named "rpi" with the given content.
func makeTarGz(t *testing.T, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: "rpi",
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("tar WriteHeader: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("tar Write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar Close: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip Close: %v", err)
	}
	return buf.Bytes()
}

func TestDownloadAndVerify_Success(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho hello\n")
	archive := makeTarGz(t, binaryContent)

	hash := sha256.Sum256(archive)
	checksumLine := fmt.Sprintf("%x  rpi_1.2.3_darwin_arm64.tar.gz\n", hash)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/archive":
			w.Write(archive)
		case "/checksums":
			w.Write([]byte(checksumLine))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	got, err := DownloadAndVerify(srv.URL+"/archive", srv.URL+"/checksums", "rpi_1.2.3_darwin_arm64.tar.gz")
	if err != nil {
		t.Fatalf("DownloadAndVerify() error: %v", err)
	}

	if !bytes.Equal(got, binaryContent) {
		t.Errorf("extracted binary content mismatch: got %q, want %q", got, binaryContent)
	}
}

func TestDownloadAndVerify_ChecksumMismatch(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho hello\n")
	archive := makeTarGz(t, binaryContent)

	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000  rpi_1.2.3_darwin_arm64.tar.gz\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/archive":
			w.Write(archive)
		case "/checksums":
			w.Write([]byte(wrongChecksum))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	_, err := DownloadAndVerify(srv.URL+"/archive", srv.URL+"/checksums", "rpi_1.2.3_darwin_arm64.tar.gz")
	if err == nil {
		t.Fatal("expected error for checksum mismatch, got nil")
	}

	if got := err.Error(); !contains(got, "checksum mismatch") {
		t.Errorf("error %q does not contain %q", got, "checksum mismatch")
	}
}

func TestFetchLatestRelease_NetworkError(t *testing.T) {
	// Use an unreachable address to trigger a network error.
	_, err := FetchLatestRelease("http://127.0.0.1:1", "A-NGJ/rpi")
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
}

func TestReplaceBinary(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "rpi")

	// Write an initial "binary".
	oldContent := []byte("old-binary")
	if err := os.WriteFile(binPath, oldContent, 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	newContent := []byte("new-binary-content")
	if err := ReplaceBinary(binPath, newContent); err != nil {
		t.Fatalf("ReplaceBinary() error: %v", err)
	}

	// Verify content was replaced.
	got, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.Equal(got, newContent) {
		t.Errorf("binary content = %q, want %q", got, newContent)
	}

	// Verify permissions.
	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0755 {
		t.Errorf("permissions = %o, want 0755", perm)
	}
}

func TestUpgrade_AlreadyUpToDate(t *testing.T) {
	release := map[string]any{
		"tag_name": "v1.0.0",
		"assets":   []map[string]any{},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	err := Upgrade(&buf, srv.URL, "A-NGJ/rpi", "1.0.0")
	if err != nil {
		t.Fatalf("Upgrade() error: %v", err)
	}

	out := buf.String()
	if !contains(out, "Already up to date") {
		t.Errorf("output %q does not contain %q", out, "Already up to date")
	}
}

func TestUpgrade_DevBuildUpgrades(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho upgraded\n")
	archive := makeTarGz(t, binaryContent)

	hash := sha256.Sum256(archive)
	archiveName := fmt.Sprintf("rpi_1.5.0_%s_%s.tar.gz", goos(), goarch())

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/A-NGJ/rpi/releases/latest":
			release := map[string]any{
				"tag_name": "v1.5.0",
				"assets": []map[string]any{
					{
						"name":                 archiveName,
						"browser_download_url": srv.URL + "/download/" + archiveName,
					},
					{
						"name":                 "checksums.txt",
						"browser_download_url": srv.URL + "/checksums/checksums.txt",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(release)
		default:
			if len(r.URL.Path) > len("/checksums/") && r.URL.Path[:len("/checksums/")] == "/checksums/" {
				checksumLine := fmt.Sprintf("%x  %s\n", hash, archiveName)
				w.Write([]byte(checksumLine))
			} else {
				w.Write(archive)
			}
		}
	}))
	defer srv.Close()

	// Set up a dummy binary that Upgrade will replace.
	dir := t.TempDir()
	dummyBin := filepath.Join(dir, "rpi")
	os.WriteFile(dummyBin, []byte("old"), 0755)

	// Override the executable path resolver so Upgrade uses our temp binary.
	origExecPath := executablePath
	executablePath = func() (string, error) { return dummyBin, nil }
	defer func() { executablePath = origExecPath }()

	var buf bytes.Buffer
	err := Upgrade(&buf, srv.URL, "A-NGJ/rpi", "dev")
	if err != nil {
		t.Fatalf("Upgrade() error: %v", err)
	}

	out := buf.String()
	if !contains(out, "v1.5.0") {
		t.Errorf("output %q does not contain new version", out)
	}

	// Verify the binary was actually replaced.
	got, err := os.ReadFile(dummyBin)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.Equal(got, binaryContent) {
		t.Errorf("binary content = %q, want %q", got, binaryContent)
	}
}

func TestUpgrade_NewerVersion(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho v2\n")
	archive := makeTarGz(t, binaryContent)

	hash := sha256.Sum256(archive)
	archiveName := fmt.Sprintf("rpi_2.0.0_%s_%s.tar.gz", goos(), goarch())

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/A-NGJ/rpi/releases/latest":
			release := map[string]any{
				"tag_name": "v2.0.0",
				"assets": []map[string]any{
					{
						"name":                 archiveName,
						"browser_download_url": srv.URL + "/download/" + archiveName,
					},
					{
						"name":                 "checksums.txt",
						"browser_download_url": srv.URL + "/checksums/checksums.txt",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(release)
		default:
			if strings.HasPrefix(r.URL.Path, "/checksums/") {
				checksumLine := fmt.Sprintf("%x  %s\n", hash, archiveName)
				w.Write([]byte(checksumLine))
			} else {
				w.Write(archive)
			}
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	dummyBin := filepath.Join(dir, "rpi")
	os.WriteFile(dummyBin, []byte("old"), 0755)

	origExecPath := executablePath
	executablePath = func() (string, error) { return dummyBin, nil }
	defer func() { executablePath = origExecPath }()

	var buf bytes.Buffer
	err := Upgrade(&buf, srv.URL, "A-NGJ/rpi", "1.0.0")
	if err != nil {
		t.Fatalf("Upgrade() error: %v", err)
	}

	out := buf.String()
	if !contains(out, "v1.0.0") || !contains(out, "v2.0.0") {
		t.Errorf("output %q should mention both old and new versions", out)
	}

	got, err := os.ReadFile(dummyBin)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.Equal(got, binaryContent) {
		t.Errorf("binary content = %q, want %q", got, binaryContent)
	}
}

// contains wraps strings.Contains for readability in test assertions.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
