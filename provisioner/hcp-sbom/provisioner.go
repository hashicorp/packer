// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package hcp_sbom

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

	"github.com/hashicorp/go-getter/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The file path or URL to the SBOM file in the Packer artifact.
	// This file must either be in the SPDX or CycloneDX format.
	// Not required if auto_generate is true.
	Source string `mapstructure:"source"`

	// The path on the local machine to store a copy of the SBOM file.
	// You can specify an absolute or a path relative to the working directory
	// when you execute the Packer build. If the file already exists on the
	// local machine, Packer overwrites the file. If the destination is a
	// directory, the directory must already exist.
	Destination string `mapstructure:"destination"`

	// The name of the SBOM file stored in HCP Packer.
	// If omitted, HCP Packer uses the build fingerprint as the file name.
	// This value must be between three and 36 characters from the following set: `[A-Za-z0-9_-]`.
	// You must specify a unique name for each build in an artifact version.
	SbomName string `mapstructure:"sbom_name"`
	
	// Native SBOM generation configuration
	// Enable automatic SBOM generation by downloading and running a scanner tool on the remote host
	AutoGenerate bool `mapstructure:"auto_generate"`
	
	// URL to scanner tool (supports go-getter syntax: HTTP, local files, Git, S3, etc.)
	// If empty and auto_generate is true, Syft will be auto-downloaded based on detected OS/Arch
	ScannerURL string `mapstructure:"scanner_url"`
	
	// Expected SHA256 checksum of scanner binary for verification
	// If provided, scanner_url must also be specified
	ScannerChecksum string `mapstructure:"scanner_checksum"`
	
	// Arguments to pass to the scanner tool
	// Default for Syft: ["-o", "spdx-json"]
	ScannerArgs []string `mapstructure:"scanner_args"`
	
	// Path to scan on remote host
	// Default: "/"
	ScanPath string `mapstructure:"scan_path"`
	
	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

var sbomFormatRegexp = regexp.MustCompile("^[0-9A-Za-z-]{3,36}$")

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "hcp-sbom",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	var errs error

	// Validate based on mode
	if p.config.AutoGenerate {
		// Native generation mode: source is optional
		// Set defaults
		if p.config.ScanPath == "" {
			p.config.ScanPath = "/"
		}
		if len(p.config.ScannerArgs) == 0 {
			p.config.ScannerArgs = []string{"-o", "spdx-json"}
		}
		
		// Validate: if checksum is provided, URL must also be provided
		if p.config.ScannerChecksum != "" && p.config.ScannerURL == "" {
			errs = packersdk.MultiErrorAppend(errs, errors.New("scanner_checksum requires scanner_url to be specified"))
		}
	} else {
		// Traditional mode: source is required
		if p.config.Source == "" {
			errs = packersdk.MultiErrorAppend(errs, errors.New("source must be specified (or enable auto_generate)"))
		}
	}

	if p.config.SbomName != "" && !sbomFormatRegexp.MatchString(p.config.SbomName) {
		// Ugly but a bit of a problem with interpolation since Provisioners
		// are prepared twice in HCL2.
		//
		// If the information used for interpolating is populated in-between the
		// first call to Prepare (at the start of the build), and when the
		// Provisioner is actually called, the first call will fail, as
		// the value won't contain the actual interpolated value, but a
		// placeholder which doesn't match the regex.
		//
		// Since we don't have a way to discriminate between the calls
		// in the context of the provisioner, we ignore them, and later the
		// HCP Packer call will fail because of the broken regex.
		if strings.Contains(p.config.SbomName, "<no value>") {
			log.Printf("[WARN] interpolation incomplete for `sbom_name`, will possibly retry later with data populated into context, otherwise will fail when uploading to HCP Packer.")
		} else {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("`sbom_name` %q doesn't match the expected format, it must "+
				"contain between 3 and 36 characters, all from the following set: [A-Za-z0-9_-]", p.config.SbomName))
		}
	}

	return errs
}

// PackerSBOM is the type we write to the temporary JSON dump of the SBOM to
// be consumed by Packer core
type PackerSBOM struct {
	// RawSBOM is the raw data from the SBOM downloaded from the guest
	RawSBOM []byte `json:"raw_sbom"`
	// Format is the format detected by the provisioner
	//
	// Supported values: `SPDX` or `CYCLONEDX`
	Format hcpPackerModels.HashicorpCloudPacker20230101SbomFormat `json:"format"`
	// Name is the name of the SBOM to be set on HCP Packer
	//
	// If unset, HCP Packer will generate one
	Name string `json:"name,omitempty"`
}

func (p *Provisioner) Provision(
	ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator,
	generatedData map[string]interface{},
) error {
	log.Println("Starting to provision with `hcp-sbom` provisioner")

	if generatedData == nil {
		generatedData = make(map[string]interface{})
	}
	p.config.ctx.Data = generatedData

	// Check if native generation is enabled
	if !p.config.AutoGenerate {
		// Original behavior: user provides SBOM file
		ui.Say("Using existing SBOM provisioner behavior (user-provided SBOM)")
		return p.provisionWithExistingSBOM(ctx, ui, comm, generatedData)
	}

	// Native generation enabled
	ui.Say("Native SBOM generation enabled")

	var osType, osArch string
	var err error

	// Only detect OS if scanner_url is NOT provided
	if p.config.ScannerURL == "" {
		ui.Say("No scanner URL provided, detecting remote OS/Arch...")
		osType, osArch, err = p.detectRemoteOS(ctx, ui, comm, generatedData)
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to detect remote OS: %s", err))
			ui.Error("SBOM generation will be skipped, but build will continue")
			return nil
		}
		ui.Say(fmt.Sprintf("Detected: OS=%s, Arch=%s", osType, osArch))
	} else {
		ui.Say("Scanner URL provided, skipping OS detection")
		// User provided scanner URL, assume they know their platform
		osType = "unknown"
		osArch = "unknown"
	}

	err = p.provisionWithNativeGeneration(ctx, ui, comm, generatedData, osType, osArch)
	if err != nil {
		ui.Error(fmt.Sprintf("SBOM generation failed: %s", err))
		ui.Error("Build will continue without SBOM")
		return nil
	}
	return nil
}

// provisionWithExistingSBOM handles the original flow where user provides an SBOM file
func (p *Provisioner) provisionWithExistingSBOM(
	ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator,
	generatedData map[string]interface{},
) error {
	src := p.config.Source

	// Download SBOM from remote
	var buf bytes.Buffer
	if err := comm.Download(src, &buf); err != nil {
		ui.Error(fmt.Sprintf("Failed to download SBOM file: %s", err))
		ui.Error("Build will continue without SBOM")
		return nil
	}

	// Process and write SBOM (reuses common logic)
	err := p.processSBOMForHCP(generatedData, buf.Bytes())
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to process SBOM: %s", err))
		ui.Error("Build will continue without SBOM")
		return nil
	}
	return nil
}

// detectRemoteOS performs OS detection on the remote host
func (p *Provisioner) detectRemoteOS(ctx context.Context, ui packersdk.Ui,
	comm packersdk.Communicator,
	generatedData map[string]interface{}) (string, string, error) {
	// First check if already detected (from generatedData)
	if osType, ok := generatedData["OSType"].(string); ok {
		if osArch, ok := generatedData["OSArch"].(string); ok {
			ui.Say("Using previously detected OS information from generated data")
			return osType, osArch, nil
		}
	}

	// Not in generatedData, detect now
	ui.Say("Running OS detection commands on remote host...")

	// Get communicator type
	connType := "ssh" // default
	if ct, ok := generatedData["ConnType"].(string); ok {
		connType = ct
	}

	// Run detection command based on communicator
	var cmd *packersdk.RemoteCmd
	if connType == "winrm" {
		cmd = &packersdk.RemoteCmd{
			Command: "echo %PROCESSOR_ARCHITECTURE%",
		}
	} else {
		cmd = &packersdk.RemoteCmd{
			Command: "uname -s -m",
		}
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := comm.Start(ctx, cmd); err != nil {
		return "", "", fmt.Errorf("failed to run OS detection command: %s", err)
	}

	cmd.Wait()

	if cmd.ExitStatus() != 0 {
		return "", "", fmt.Errorf("OS detection command exited with status %d", cmd.ExitStatus())
	}

	output := strings.TrimSpace(stdout.String())
	ui.Say(fmt.Sprintf("OS detection output: %s", output))

	// Parse output
	var osType, osArch string
	if connType == "winrm" {
		osType = "Windows"
		osArch = strings.ToLower(output) // AMD64, ARM64, etc.
	} else {
		parts := strings.Fields(output)
		if len(parts) >= 2 {
			osType = parts[0] // Linux, Darwin, FreeBSD, etc.
			osArch = parts[1] // x86_64, aarch64, etc.
		} else if len(parts) == 1 {
			// Some systems might only return one value
			osType = parts[0]
			osArch = "unknown"
		}
	}

	if osType == "" || osArch == "" {
		return "", "", fmt.Errorf("failed to parse OS detection output: %s", output)
	}

	// Store in generatedData for potential reuse
	generatedData["OSType"] = osType
	generatedData["OSArch"] = osArch

	return osType, osArch, nil
}

// getUserDestination determines and returns the destination path for the user SBOM file.
func (p *Provisioner) getUserDestination() (string, error) {
	dst := p.config.Destination

	// Check if the destination exists and determine its type
	info, err := os.Stat(dst)
	if err == nil {
		if info.IsDir() {
			// If the destination is a directory, create a temporary file inside it
			tmpFile, err := os.CreateTemp(dst, "packer-user-sbom-*.json")
			if err != nil {
				return "", fmt.Errorf("failed to create temporary file in user SBOM directory %s: %s", dst, err)
			}
			dst = tmpFile.Name()
			tmpFile.Close()
		}
		return dst, nil
	}

	outDir := filepath.Dir(dst)
	// In case the destination does not exist, we'll get the dirpath,
	// and create it if it doesn't already exist
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create destination directory for user SBOM: %s\n", err)
	}

	// Check if the destination is a directory after the previous step.
	//
	// This happens if the path specified ends with a `/`, in which case the
	// destination is a directory, and we must create a temporary file in
	// this destination directory.
	destStat, statErr := os.Stat(dst)
	if statErr == nil && destStat.IsDir() {
		tmpFile, err := os.CreateTemp(outDir, "packer-user-sbom-*.json")
		if err != nil {
			return "", fmt.Errorf("failed to create temporary file in user SBOM directory %s: %s", dst, err)
		}
		dst = tmpFile.Name()
		tmpFile.Close()
	}

	return dst, nil
}


// provisionWithNativeGeneration handles the new native SBOM generation flow
func (p *Provisioner) provisionWithNativeGeneration(
	ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator,
	generatedData map[string]interface{}, osType, osArch string,
) error {
	ui.Say("Starting native SBOM generation workflow...")

	// Step 1: Download scanner binary
	ui.Say("Downloading scanner binary...")
	scannerLocalPath, err := p.downloadScanner(ctx, ui, osType, osArch)
	if err != nil {
		return fmt.Errorf("failed to download scanner: %s", err)
	}
	defer os.Remove(scannerLocalPath)

	// Step 2: Verify checksum if provided
	if p.config.ScannerChecksum != "" {
		ui.Say("Verifying scanner checksum...")
		if err := p.verifyChecksum(scannerLocalPath); err != nil {
			return fmt.Errorf("checksum verification failed: %s", err)
		}
		ui.Say("Checksum verification passed")
	}

	// Step 3: Upload scanner to remote
	ui.Say("Uploading scanner to remote host...")
	remoteScannerPath, err := p.uploadScanner(ctx, ui, comm, scannerLocalPath, osType)
	if err != nil {
		return fmt.Errorf("failed to upload scanner: %s", err)
	}
	defer p.cleanupRemoteFile(ctx, ui, comm, remoteScannerPath)

	// Step 4: Run scanner on remote
	ui.Say(fmt.Sprintf("Running scanner on remote host (scanning %s)...", p.config.ScanPath))
	remoteSBOMPath, err := p.runScanner(ctx, ui, comm, remoteScannerPath, osType)
	if err != nil {
		return fmt.Errorf("failed to run scanner: %s", err)
	}
	defer p.cleanupRemoteFile(ctx, ui, comm, remoteSBOMPath)

	// Step 5: Download SBOM from remote
	ui.Say("Downloading SBOM from remote host...")
	sbomData, err := p.downloadSBOM(ctx, ui, comm, remoteSBOMPath)
	if err != nil {
		return fmt.Errorf("failed to download SBOM: %s", err)
	}

	// Step 6: Process SBOM for HCP (validate, compress, store)
	ui.Say("Processing SBOM for HCP Packer...")
	if err := p.processSBOMForHCP(generatedData, sbomData); err != nil {
		return fmt.Errorf("failed to process SBOM: %s", err)
	}

	ui.Say("Native SBOM generation completed successfully")
	return nil
}

// downloadScanner downloads the scanner binary using go-getter
func (p *Provisioner) downloadScanner(ctx context.Context, ui packersdk.Ui,
	osType, osArch string) (string, error) {
	var downloadURL string

	// If user provided a URL, use it
	if p.config.ScannerURL != "" {
		downloadURL = p.config.ScannerURL
		ui.Say(fmt.Sprintf("Using custom scanner URL: %s", downloadURL))
	} else {
		// Default to Syft from GitHub releases
		if osType == "unknown" || osArch == "unknown" {
			return "", fmt.Errorf("cannot auto-download scanner: OS/Arch unknown (provide scanner_url)")
		}
		ui.Say(fmt.Sprintf("Fetching latest Syft version for %s/%s...", osType, osArch))
		downloadURL = p.buildDefaultSyftURL(osType, osArch)
		ui.Say(fmt.Sprintf("Download URL: %s", downloadURL))
	}

	// Create temporary directory for download
	tmpDir, err := os.MkdirTemp("", "packer-scanner-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use go-getter to download
	client := &getter.Client{}
	
	req := &getter.Request{
		Src: downloadURL,
		Dst: tmpDir,
	}

	ui.Say("Downloading scanner binary...")
	if _, err := client.Get(ctx, req); err != nil {
		return "", fmt.Errorf("failed to download scanner: %s", err)
	}

	// Find the scanner binary in the downloaded files
	scannerPath, err := p.findScannerBinary(tmpDir, osType)
	if err != nil {
		return "", fmt.Errorf("failed to locate scanner binary: %s", err)
	}

	// Copy to a permanent temp location
	finalPath, err := p.copyScannerToTemp(scannerPath)
	if err != nil {
		return "", fmt.Errorf("failed to copy scanner: %s", err)
	}

	ui.Say(fmt.Sprintf("Scanner downloaded to: %s", finalPath))
	return finalPath, nil
}

// buildDefaultSyftURL constructs the default Syft download URL
func (p *Provisioner) buildDefaultSyftURL(osType, osArch string) string {
	// Map to Syft platform naming
	syftOS, syftArch := p.mapToSyftPlatform(osType, osArch)

	// Fetch latest version from GitHub API
	version := p.getLatestSyftVersion()
	if version == "" {
		// Fallback to a known stable version if API call fails
		log.Printf("[WARN] Failed to fetch latest Syft version, using fallback v0.100.0")
		version = "v0.100.0"
	}

	// Construct GitHub release URL
	// Example: https://github.com/anchore/syft/releases/download/v0.100.0/syft_0.100.0_linux_amd64.tar.gz
	versionNum := strings.TrimPrefix(version, "v")
	fileName := fmt.Sprintf("syft_%s_%s_%s.tar.gz", versionNum, syftOS, syftArch)

	return fmt.Sprintf("https://github.com/anchore/syft/releases/download/%s/%s",
		version, fileName)
}

// getLatestSyftVersion fetches the latest Syft release version from GitHub API
func (p *Provisioner) getLatestSyftVersion() string {
	// GitHub API endpoint for latest release
	apiURL := "https://api.github.com/repos/anchore/syft/releases/latest"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("[WARN] Failed to create request for Syft version: %s", err)
		return ""
	}

	// Set User-Agent header (GitHub API requires it)
	req.Header.Set("User-Agent", "Packer-HCP-SBOM-Provisioner")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[WARN] Failed to fetch latest Syft version: %s", err)
		return ""
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("[WARN] GitHub API returned status %d for Syft version", resp.StatusCode)
		return ""
	}

	// Parse response
	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.Printf("[WARN] Failed to parse Syft version response: %s", err)
		return ""
	}

	if release.TagName == "" {
		log.Printf("[WARN] Empty tag_name in Syft release response")
		return ""
	}

	log.Printf("[INFO] Latest Syft version: %s", release.TagName)
	return release.TagName
}

// mapToSyftPlatform maps detected OS/Arch to Syft naming conventions
func (p *Provisioner) mapToSyftPlatform(osType, osArch string) (string, string) {
	osType = strings.ToLower(osType)
	osArch = strings.ToLower(osArch)

	// Map OS
	syftOS := "linux"
	if strings.Contains(osType, "darwin") || strings.Contains(osType, "macos") {
		syftOS = "darwin"
	} else if strings.Contains(osType, "windows") {
		syftOS = "windows"
	} else if strings.Contains(osType, "freebsd") {
		syftOS = "freebsd"
	}

	// Map Architecture
	syftArch := osArch
	switch osArch {
	case "x86_64", "amd64":
		syftArch = "amd64"
	case "aarch64", "arm64":
		syftArch = "arm64"
	case "i386", "i686":
		syftArch = "386"
	case "armv7l", "armv7":
		syftArch = "arm"
	}

	return syftOS, syftArch
}

// findScannerBinary locates the scanner executable in the downloaded directory
func (p *Provisioner) findScannerBinary(dir, osType string) (string, error) {
	osType = strings.ToLower(osType)

	var binaryName string
	if strings.Contains(osType, "windows") {
		binaryName = "syft.exe"
	} else {
		binaryName = "syft"
	}

	var foundPath string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileName := filepath.Base(path)
			// Match exact name or name as part of the file
			if fileName == binaryName || strings.Contains(fileName, binaryName) {
				// For archives, we want the actual binary, not the archive
				if !strings.HasSuffix(fileName, ".tar.gz") &&
					!strings.HasSuffix(fileName, ".zip") &&
					!strings.HasSuffix(fileName, ".tar") {
					foundPath = path
					return filepath.SkipDir
				}
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if foundPath == "" {
		// If not found, try to extract from tar.gz
		foundPath, err = p.extractScannerFromArchive(dir, binaryName)
		if err != nil {
			return "", fmt.Errorf("scanner binary '%s' not found in downloaded files", binaryName)
		}
	}

	return foundPath, nil
}

// extractScannerFromArchive extracts the scanner binary from a tar.gz archive
func (p *Provisioner) extractScannerFromArchive(dir, binaryName string) (string, error) {
	// Find tar.gz file
	var archivePath string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tar.gz") {
			archivePath = path
			return filepath.SkipDir
		}
		return nil
	})

	if archivePath == "" {
		return "", fmt.Errorf("no tar.gz archive found")
	}

	// Open and extract
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Find and extract the binary
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the binary
		if filepath.Base(header.Name) == binaryName {
			// Create temporary file for the binary
			tmpBinary, err := os.CreateTemp(dir, "syft-binary-*")
			if err != nil {
				return "", err
			}
			defer tmpBinary.Close()

			// Copy binary content
			if _, err := io.Copy(tmpBinary, tr); err != nil {
				return "", err
			}

			// Make executable
			if err := os.Chmod(tmpBinary.Name(), 0755); err != nil {
				return "", err
			}

			return tmpBinary.Name(), nil
		}
	}

	return "", fmt.Errorf("binary '%s' not found in archive", binaryName)
}

// copyScannerToTemp copies the scanner binary to a permanent temp location
func (p *Provisioner) copyScannerToTemp(srcPath string) (string, error) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "packer-scanner-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Open source
	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Copy
	if _, err := io.Copy(tmpFile, src); err != nil {
		return "", err
	}

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// verifyChecksum verifies the SHA256 checksum of the scanner binary
func (p *Provisioner) verifyChecksum(filePath string) error {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Calculate SHA256
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	// Get hex string
	actualChecksum := hex.EncodeToString(hash.Sum(nil))

	// Compare with expected
	expectedChecksum := strings.ToLower(strings.TrimSpace(p.config.ScannerChecksum))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}


// uploadScanner uploads the scanner binary to the remote host
func (p *Provisioner) uploadScanner(ctx context.Context, ui packersdk.Ui,
	comm packersdk.Communicator, localPath, osType string) (string, error) {
	
	// Determine remote path based on OS
	var remotePath string
	if strings.Contains(strings.ToLower(osType), "windows") {
		remotePath = "C:\\Windows\\Temp\\packer-sbom-scanner.exe"
	} else {
		remotePath = "/tmp/packer-sbom-scanner"
	}

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open local scanner: %s", err)
	}
	defer localFile.Close()

	// Upload to remote
	ui.Say(fmt.Sprintf("Uploading scanner to %s...", remotePath))
	if err := comm.Upload(remotePath, localFile, nil); err != nil {
		return "", fmt.Errorf("failed to upload scanner: %s", err)
	}

	// Make executable on Unix-like systems
	if !strings.Contains(strings.ToLower(osType), "windows") {
		cmd := &packersdk.RemoteCmd{
			Command: fmt.Sprintf("chmod +x %s", remotePath),
		}
		if err := comm.Start(ctx, cmd); err != nil {
			return "", fmt.Errorf("failed to make scanner executable: %s", err)
		}
		cmd.Wait()
		if cmd.ExitStatus() != 0 {
			return "", fmt.Errorf("chmod command failed with exit status %d", cmd.ExitStatus())
		}
	}

	return remotePath, nil
}

// runScanner executes the scanner on the remote host
func (p *Provisioner) runScanner(ctx context.Context, ui packersdk.Ui,
	comm packersdk.Communicator, scannerPath, osType string) (string, error) {
	
	// Determine output path based on OS
	var outputPath string
	if strings.Contains(strings.ToLower(osType), "windows") {
		outputPath = "C:\\Windows\\Temp\\packer-sbom.json"
	} else {
		outputPath = "/tmp/packer-sbom.json"
	}

	// Build scanner command
	args := append(p.config.ScannerArgs, p.config.ScanPath)
	
	// Add output redirection
	var cmdStr string
	if strings.Contains(strings.ToLower(osType), "windows") {
		cmdStr = fmt.Sprintf("%s %s > %s", scannerPath, strings.Join(args, " "), outputPath)
	} else {
		cmdStr = fmt.Sprintf("sudo %s %s > %s", scannerPath, strings.Join(args, " "), outputPath)
	}

	ui.Say(fmt.Sprintf("Executing: %s", cmdStr))

	// Execute scanner
	var stdout, stderr bytes.Buffer
	cmd := &packersdk.RemoteCmd{
		Command: cmdStr,
		Stdout:  &stdout,
		Stderr:  &stderr,
	}

	if err := comm.Start(ctx, cmd); err != nil {
		return "", fmt.Errorf("failed to start scanner: %s", err)
	}

	cmd.Wait()

	// Log output
	if stdout.Len() > 0 {
		ui.Say(fmt.Sprintf("Scanner stdout: %s", stdout.String()))
	}
	if stderr.Len() > 0 {
		ui.Say(fmt.Sprintf("Scanner stderr: %s", stderr.String()))
	}

	if cmd.ExitStatus() != 0 {
		return "", fmt.Errorf("scanner exited with status %d", cmd.ExitStatus())
	}

	return outputPath, nil
}

// downloadSBOM downloads the SBOM file from the remote host
func (p *Provisioner) downloadSBOM(ctx context.Context, ui packersdk.Ui,
	comm packersdk.Communicator, remotePath string) ([]byte, error) {
	
	var buf bytes.Buffer
	ui.Say(fmt.Sprintf("Downloading SBOM from %s...", remotePath))
	
	if err := comm.Download(remotePath, &buf); err != nil {
		return nil, fmt.Errorf("failed to download SBOM: %s", err)
	}

	if buf.Len() == 0 {
		return nil, fmt.Errorf("downloaded SBOM is empty")
	}

	ui.Say(fmt.Sprintf("Downloaded SBOM (%d bytes)", buf.Len()))
	return buf.Bytes(), nil
}

// cleanupRemoteFile removes a file from the remote host
func (p *Provisioner) cleanupRemoteFile(ctx context.Context, ui packersdk.Ui,
	comm packersdk.Communicator, remotePath string) {
	
	if remotePath == "" {
		return
	}

	ui.Say(fmt.Sprintf("Cleaning up remote file: %s", remotePath))
	
	// Determine delete command based on path
	var cmdStr string
	if strings.Contains(remotePath, "C:\\") || strings.Contains(remotePath, "c:\\") {
		cmdStr = fmt.Sprintf("del /F /Q %s", remotePath)
	} else {
		cmdStr = fmt.Sprintf("rm -f %s", remotePath)
	}

	cmd := &packersdk.RemoteCmd{
		Command: cmdStr,
	}

	if err := comm.Start(ctx, cmd); err != nil {
		ui.Error(fmt.Sprintf("Failed to cleanup remote file %s: %s", remotePath, err))
		return
	}

	cmd.Wait()
	if cmd.ExitStatus() != 0 {
		ui.Error(fmt.Sprintf("Cleanup command failed for %s with exit status %d", remotePath, cmd.ExitStatus()))
	}
}

// processSBOMForHCP validates, compresses, and prepares SBOM for HCP upload
func (p *Provisioner) processSBOMForHCP(generatedData map[string]interface{}, sbomData []byte) error {
	// Validate SBOM format
	format, err := validateSBOM(sbomData)
	if err != nil {
		return fmt.Errorf("SBOM validation failed: %s", err)
	}

	// Get destination path from generatedData
	pkrDst, ok := generatedData["dst"].(string)
	if !ok || pkrDst == "" {
		return fmt.Errorf("packer destination path missing from configs: this is an internal error")
	}

	// Write PackerSBOM to destination
	outFile, err := os.Create(pkrDst)
	if err != nil {
		return fmt.Errorf("failed to create output file %q: %s", pkrDst, err)
	}
	defer outFile.Close()

	err = json.NewEncoder(outFile).Encode(PackerSBOM{
		RawSBOM: sbomData,
		Format:  format,
		Name:    p.config.SbomName,
	})
	if err != nil {
		return fmt.Errorf("failed to write SBOM to %q: %s", pkrDst, err)
	}

	// Also save to user destination if specified
	if p.config.Destination != "" {
		usrDst, err := p.getUserDestination()
		if err != nil {
			return fmt.Errorf("failed to compute destination path %q: %s", p.config.Destination, err)
		}
		if err := os.WriteFile(usrDst, sbomData, 0644); err != nil {
			return fmt.Errorf("failed to write SBOM to destination %q: %s", usrDst, err)
		}
	}

	return nil
}
