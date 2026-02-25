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
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	// Not required if enable_native_generation is true.
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
	// Enable native SBOM generation by automatically downloading and running a scanner
	EnableNativeGeneration bool `mapstructure:"enable_native_generation"`
	
	// URL to scanner tool (supports go-getter syntax: HTTP, local files, Git, S3, etc.)
	// If empty and enable_native_generation is true, Syft will be auto-downloaded based on detected OS/Arch
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
	if p.config.EnableNativeGeneration {
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
			errs = packersdk.MultiErrorAppend(errs, errors.New("source must be specified (or enable enable_native_generation)"))
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
	if !p.config.EnableNativeGeneration {
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
			return fmt.Errorf("failed to detect remote OS: %s", err)
		}
		ui.Say(fmt.Sprintf("Detected: OS=%s, Arch=%s", osType, osArch))
	} else {
		ui.Say("Scanner URL provided, skipping OS detection")
		// User provided scanner URL, assume they know their platform
		osType = "unknown"
		osArch = "unknown"
	}

	return p.provisionWithNativeGeneration(ctx, ui, comm, generatedData, osType, osArch)
}

// provisionWithExistingSBOM handles the original flow where user provides an SBOM file
func (p *Provisioner) provisionWithExistingSBOM(
	ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator,
	generatedData map[string]interface{},
) error {
	src := p.config.Source

	pkrDst := generatedData["dst"].(string)
	if pkrDst == "" {
		return fmt.Errorf("packer destination path missing from configs: this is an internal error, which should be reported to be fixed.")
	}

	var buf bytes.Buffer
	if err := comm.Download(src, &buf); err != nil {
		ui.Errorf("download failed for SBOM file: %s", err)
		return err
	}

	format, err := validateSBOM(buf.Bytes())
	if err != nil {
		return fmt.Errorf("validation failed for SBOM file: %s", err)
	}

	outFile, err := os.Create(pkrDst)
	if err != nil {
		return fmt.Errorf("failed to open/create output file %q: %s", pkrDst, err)
	}
	defer outFile.Close()

	err = json.NewEncoder(outFile).Encode(PackerSBOM{
		RawSBOM: buf.Bytes(),
		Format:  format,
		Name:    p.config.SbomName,
	})
	if err != nil {
		return fmt.Errorf("failed to write sbom file to %q: %s", pkrDst, err)
	}

	if p.config.Destination == "" {
		return nil
	}

	// SBOM for User
	usrDst, err := p.getUserDestination()
	if err != nil {
		return fmt.Errorf("failed to compute destination path %q: %s", p.config.Destination, err)
	}
	err = os.WriteFile(usrDst, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write SBOM to destination %q: %s", usrDst, err)
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
			ui.Message("Using previously detected OS information from generated data")
			return osType, osArch, nil
		}
	}

	// Not in generatedData, detect now
	ui.Message("Running OS detection commands on remote host...")

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
	ui.Message(fmt.Sprintf("OS detection output: %s", output))

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
	// Step 4: Run scanner on remote
	// Step 5: Download SBOM
	// Step 6: Process SBOM for HCP
	// Step 7: Cleanup remote files
	// TODO: Implement in commits 4-6

	return fmt.Errorf("native SBOM generation partially implemented (commits 4-6 pending)")
}

// downloadScanner downloads the scanner binary using go-getter
func (p *Provisioner) downloadScanner(ctx context.Context, ui packersdk.Ui,
	osType, osArch string) (string, error) {
	var downloadURL string

	// If user provided a URL, use it
	if p.config.ScannerURL != "" {
		downloadURL = p.config.ScannerURL
		ui.Message(fmt.Sprintf("Using custom scanner URL: %s", downloadURL))
	} else {
		// Default to Syft from GitHub releases
		if osType == "unknown" || osArch == "unknown" {
			return "", fmt.Errorf("cannot auto-download scanner: OS/Arch unknown (provide scanner_url)")
		}
		downloadURL = p.buildDefaultSyftURL(osType, osArch)
		ui.Message(fmt.Sprintf("Auto-downloading Syft for %s/%s", osType, osArch))
		ui.Message(fmt.Sprintf("Download URL: %s", downloadURL))
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

	ui.Message("Downloading scanner binary...")
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

	ui.Message(fmt.Sprintf("Scanner downloaded to: %s", finalPath))
	return finalPath, nil
}

// buildDefaultSyftURL constructs the default Syft download URL
func (p *Provisioner) buildDefaultSyftURL(osType, osArch string) string {
	// Map to Syft platform naming
	syftOS, syftArch := p.mapToSyftPlatform(osType, osArch)

	// Default to latest stable version
	version := "v0.100.0"

	// Construct GitHub release URL
	// Example: https://github.com/anchore/syft/releases/download/v0.100.0/syft_0.100.0_linux_amd64.tar.gz
	versionNum := strings.TrimPrefix(version, "v")
	fileName := fmt.Sprintf("syft_%s_%s_%s.tar.gz", versionNum, syftOS, syftArch)

	return fmt.Sprintf("https://github.com/anchore/syft/releases/download/%s/%s",
		version, fileName)
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
