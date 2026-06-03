// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package hcp_sbom

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	semver "github.com/Masterminds/semver/v3"
	"github.com/hashicorp/packer-plugin-sdk/retry"
)

// releaseBaseURL is the base URL for downloading Packer release artifacts.
const defaultReleaseBaseURL = "https://releases.hashicorp.com"

func getReleaseBaseURL() string {
	return defaultReleaseBaseURL
}

// releaseIndex is the top-level structure of https://releases.hashicorp.com/packer/index.json.
type releaseIndex struct {
	Versions map[string]releaseVersion `json:"versions"`
}

// releaseVersion represents one version entry in the release index.
type releaseVersion struct {
	Version string         `json:"version"`
	Shasums string         `json:"shasums"`
	Builds  []releaseBuild `json:"builds"`
}

// releaseBuild represents one platform build inside a release version.
type releaseBuild struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

// fetchLatestPackerVersion queries the HashiCorp releases index, sorts all
// stable (non-prerelease) versions with semver, and returns the highest one.
func fetchLatestPackerVersion(ctx context.Context, client *http.Client) (string, error) {
	indexURL := getReleaseBaseURL() + "/packer/index.json"
	var indexData releaseIndex

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, indexURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build index request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch release index: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d for %s", resp.StatusCode, indexURL)
	}

	err = json.NewDecoder(resp.Body).Decode(&indexData)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve packer release index from %s: %w", indexURL, err)
	}

	var semverList []*semver.Version
	for vStr := range indexData.Versions {
		v, parseErr := semver.NewVersion(vStr)
		if parseErr != nil {
			continue
		}
		if v.Prerelease() != "" {
			continue // skip alpha/beta/rc
		}
		semverList = append(semverList, v)
	}

	if len(semverList) == 0 {
		return "", fmt.Errorf("no stable Packer releases found in index at %s", indexURL)
	}

	sort.Sort(semver.Collection(semverList))
	latest := semverList[len(semverList)-1]
	log.Printf("[INFO] Latest stable Packer version from releases index: %s", latest.Original())
	return latest.Original(), nil
}

// downloadURLToTempFile downloads url into a new temp file and returns its path.
// On any error the temp file is removed. The caller owns the returned file on success.
func downloadURLToTempFile(ctx context.Context, client *http.Client, url, suffix string) (string, error) {
	f, err := os.CreateTemp("", "packer-dl-*"+suffix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := f.Name()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	_, copyErr := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("failed to write download: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("failed to close temp file: %w", closeErr)
	}

	return tmpPath, nil
}

// downloadChecksumFile fetches the SHA256SUMS text file at url.
func downloadChecksumFile(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build request for %s: %w", url, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed reading response body for %s: %w", url, err)
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return "", fmt.Errorf("empty response body for %s", url)
	}

	return string(body), nil
}

func isValidSHA256Hex(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func expectedZipSHA256FromSums(sumsContent, fileName string) (string, error) {
	for _, line := range strings.Split(sumsContent, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 {
			continue
		}
		candidateFileName := strings.TrimPrefix(fields[len(fields)-1], "*")
		if candidateFileName == fileName {
			hash := strings.ToLower(fields[0])
			if !isValidSHA256Hex(hash) {
				return "", fmt.Errorf("invalid SHA256 checksum format for %s in SHA256SUMS", fileName)
			}
			return hash, nil
		}
	}
	return "", fmt.Errorf("checksum for %s not found in SHA256SUMS", fileName)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open %s for hashing: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed hashing %s: %w", path, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// downloadPackerRelease fetches the latest stable Packer version from the
// HashiCorp releases index (releases.hashicorp.com/packer/index.json), then
// downloads and checksum-verifies the zip for the given GOOS/GOARCH.
// All HTTP operations are retried up to three times.
func downloadPackerRelease(ctx context.Context, goos, goarch string) (string, error) {
	base := getReleaseBaseURL()
	client := &http.Client{Timeout: 5 * time.Minute}

	var zipPath string
	err := retry.Config{
		Tries:      3,
		RetryDelay: func() time.Duration { return 5 * time.Second },
	}.Run(ctx, func(ctx context.Context) error {
		// Resolve the latest stable version from the releases index.
		v, err := fetchLatestPackerVersion(ctx, client)
		if err != nil {
			return fmt.Errorf("failed to determine latest Packer version: %w", err)
		}

		fileName := fmt.Sprintf("packer_%s_%s_%s.zip", v, goos, goarch)
		zipURL := fmt.Sprintf("%s/packer/%s/%s", base, v, fileName)
		shaSumsURL := fmt.Sprintf("%s/packer/%s/packer_%s_SHA256SUMS", base, v, v)

		log.Printf("[INFO] Downloading and verifying Packer %s for %s/%s...", v, goos, goarch)

		candidateZipPath, err := downloadURLToTempFile(ctx, client, zipURL, ".zip")
		if err != nil {
			return fmt.Errorf("failed to download Packer release zip: %w", err)
		}
		keepCandidate := false
		defer func() {
			if !keepCandidate {
				_ = os.Remove(candidateZipPath)
			}
		}()

		sumsContent, err := downloadChecksumFile(ctx, client, shaSumsURL)
		if err != nil {
			return fmt.Errorf("failed to download release checksums: %w", err)
		}

		expectedSHA, err := expectedZipSHA256FromSums(sumsContent, fileName)
		if err != nil {
			return fmt.Errorf("failed to resolve expected checksum: %w", err)
		}

		actualSHA, err := fileSHA256(candidateZipPath)
		if err != nil {
			return err
		}

		if !strings.EqualFold(expectedSHA, actualSHA) {
			return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", fileName, expectedSHA, actualSHA)
		}

		// Validate the expected binary exists inside the archive.
		binaryName := "packer"
		if goos == "windows" {
			binaryName = "packer.exe"
		}

		zr, err := zip.OpenReader(candidateZipPath)
		if err != nil {
			return fmt.Errorf("failed to open downloaded zip: %w", err)
		}
		defer func() { _ = zr.Close() }()

		foundBinary := false
		for _, f := range zr.File {
			if f.Name == binaryName {
				foundBinary = true
				break
			}
		}
		if !foundBinary {
			return fmt.Errorf("packer binary %q not found in release zip %s", binaryName, zipURL)
		}

		keepCandidate = true
		zipPath = candidateZipPath
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to download and verify Packer release zip: %w", err)
	}

	log.Printf("[INFO] Downloaded and verified Packer release zip: %s", zipPath)
	return zipPath, nil
}
