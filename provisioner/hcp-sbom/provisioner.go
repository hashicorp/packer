// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package hcp_sbom

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/guestexec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The file path or URL to the SBOM file in the Packer artifact.
	// This file must either be in the SPDX or CycloneDX format.
	// Mutually exclusive with `auto_generate`.
	Source string `mapstructure:"source" required:"true"`

	// The path on the local machine to store a copy of the SBOM file.
	// You can specify an absolute or a path relative to the working directory
	// when you execute the Packer build. If the file already exists on the
	// local machine, Packer overwrites the file. If the destination is a
	// directory, the directory must already exist.
	Destination string `mapstructure:"destination" required:"false"`

	// The name of the SBOM file stored in HCP Packer.
	// If omitted, HCP Packer uses the build fingerprint as the file name.
	// This value must be between three and 36 characters from the following
	// set: `[A-Za-z0-9_-]`. You must specify a unique name for each build in
	// an artifact version.
	SbomName string `mapstructure:"sbom_name" required:"false"`

	// Enable automatic SBOM generation by running `packer sbom-generate` on
	// the remote host. When enabled, the provisioner uploads the running Packer
	// binary (which embeds the Syft SDK) to the remote VM and executes it there
	// to generate an SBOM. Mutually exclusive with `source`.
	AutoGenerate bool `mapstructure:"auto_generate" required:"true"`

	// Arguments to pass to `packer sbom-generate`. Default:
	// `["-o", "cyclonedx-json"]`.
	ScannerArgs []string `mapstructure:"scanner_args" required:"false"`

	// DEPRECATED: Custom scanner URL is no longer supported. The hcp-sbom
	// provisioner now uses the Packer binary with embedded Syft SDK for
	// automatic SBOM generation. This field is ignored and will be removed
	// in a future major version. For custom SBOM tools, use manual generation
	// with the `source` field instead of `auto_generate`.
	ScannerURL string `mapstructure:"scanner_url" required:"false"`

	// DEPRECATED: Scanner checksum verification is no longer supported.
	// This field is ignored and will be removed in a future major version.
	ScannerChecksum string `mapstructure:"scanner_checksum" required:"false"`

	// Path to scan on remote host. Defaults to `/` (root directory).
	ScanPath string `mapstructure:"scan_path" required:"false"`

	// The command template used to execute the scanner on the remote host.
	// Available template variables:
	//
	// - `{{.Path}}` - Path to the scanner binary on the remote host
	// - `{{.Args}}` - Scanner arguments (from `scanner_args`)
	// - `{{.ScanPath}}` - Path to scan (from `scan_path`)
	// - `{{.Output}}` - Output file path for the SBOM
	//
	// Default for Unix: `chmod +x {{.Path}} && sudo {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}`
	//
	// Default for Windows: `{{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}`
	//
	// Examples:
	//
	// Without sudo:
	//
	// ``` hcl
	// execute_command = "chmod +x {{.Path}} && {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}"
	// ```
	//
	// With sudo password:
	//
	// ``` hcl
	// execute_command = "chmod +x {{.Path}} && echo 'password' | sudo -S {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}"
	// ```
	ExecuteCommand string `mapstructure:"execute_command" required:"false"`

	// A username to use for elevated permissions when running Packer on
	// Windows. This is only used for Windows hosts when elevated privileges
	// are required. For Unix-like systems, use `execute_command` with sudo instead.
	ElevatedUser string `mapstructure:"elevated_user" required:"false"`

	// The password for the `elevated_user`. Required if `elevated_user` is
	// specified. Only applicable for Windows hosts.
	ElevatedPassword string `mapstructure:"elevated_password" required:"false"`

	ctx interpolate.Context
}

type Provisioner struct {
	config        Config
	communicator  packersdk.Communicator
	generatedData map[string]any
}

func formatUIWarning(message string) string {
	if os.Getenv("PACKER_NO_COLOR") != "" {
		return "WARNING: " + message
	}
	return "\033[33mWARNING:\033[0m " + message
}

func (p *Provisioner) warnDeprecatedConfigInUI(ui packersdk.Ui) {
	if p.config.ScannerURL != "" {
		ui.Say(formatUIWarning("'scanner_url' is deprecated and ignored. This field will be removed in a future version."))
	}
	if p.config.ScannerChecksum != "" {
		ui.Say(formatUIWarning("'scanner_checksum' is deprecated and ignored. This field will be removed in a future version."))
	}
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

func (p *Provisioner) FlatConfig() any {
	return p.config.FlatMapstructure()
}

var sbomFormatRegexp = regexp.MustCompile("^[0-9A-Za-z-]{3,36}$")

// scannerPathTokenRegexp matches the raw execute_command template token used
// for the uploaded binary path, including optional whitespace inside the
// template braces.
//
// Examples that match:
//
//	{{.Path}}
//	{{ .Path }}
//
// Examples that do not match:
//
//	{{.Args}}
//	/tmp/packer-sbom-runner
var scannerPathTokenRegexp = regexp.MustCompile(`\{\{\s*\.Path\s*\}\}`)

// scannerArgsOrScanPathTokenPrefixRegexp matches only when the next
// non-whitespace token after {{.Path}} is either {{.Args}} or {{.ScanPath}}.
// This is the backward-compatible shape of older scanner commands where the
// path was executed directly without an explicit sbom-generate subcommand.
//
// Examples that match after trimming leading whitespace:
//
//	{{.Args}} {{.ScanPath}} > {{.Output}}
//	{{ .ScanPath }} > {{.Output}}
//
// Examples that do not match:
//
//	sbom-generate {{.Args}} {{.ScanPath}}
//	version
//	&& chmod +x {{.Path}}
var scannerArgsOrScanPathTokenPrefixRegexp = regexp.MustCompile(`^\{\{\s*\.(Args|ScanPath)\s*\}\}`)

func (p *Provisioner) Prepare(raws ...any) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "hcp-sbom",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	var errs error

	// Validate that source and auto_generate are mutually exclusive
	if p.config.Source != "" && p.config.AutoGenerate {
		errs = packersdk.MultiErrorAppend(errs, errors.New("source and auto_generate are mutually exclusive; use either source for pre-generated SBOMs or auto_generate to create them"))
	}

	// Validate based on mode
	if p.config.AutoGenerate {
		// Native generation mode: source must not be set
		// Set defaults
		if p.config.ScanPath == "" {
			p.config.ScanPath = "/"
		}
		if len(p.config.ScannerArgs) == 0 {
			// Default to CycloneDX JSON format
			p.config.ScannerArgs = []string{
				"-o", "cyclonedx-json",
			}
		}

		// Set default execute_command if not provided
		// Note: This will be further customized based on OS at runtime
		if p.config.ExecuteCommand == "" {
			p.config.ExecuteCommand = "chmod +x {{.Path}} && sudo {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}"
		}

		// Keep legacy validation for clarity while fields remain accepted.
		if p.config.ScannerChecksum != "" && p.config.ScannerURL == "" {
			errs = packersdk.MultiErrorAppend(errs, errors.New("scanner_checksum requires scanner_url to be specified (note: both fields are deprecated and ignored)"))
		}

		// Validate elevated user configuration (Windows only)
		if p.config.ElevatedUser == "" && p.config.ElevatedPassword != "" {
			errs = packersdk.MultiErrorAppend(errs, errors.New("elevated_user must be specified if elevated_password is provided"))
		}
	} else {
		// Traditional mode: source is required
		if p.config.Source == "" {
			errs = packersdk.MultiErrorAppend(errs, errors.New("source must be specified when auto_generate is not enabled"))
		}

		// Note: Scanner-related fields are allowed in source mode to support
		// toggling auto_generate without clearing configuration fields
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
	generatedData map[string]any,
) error {
	// Store communicator and generatedData for elevated execution
	p.communicator = comm
	p.generatedData = generatedData
	log.Println("Starting to provision with `hcp-sbom` provisioner")

	if generatedData == nil {
		generatedData = make(map[string]any)
	}
	p.config.ctx.Data = generatedData

	// Check if native generation is enabled
	if !p.config.AutoGenerate {
		// Original behavior: user provides SBOM file
		log.Println("Using existing SBOM provisioner behavior (user-provided SBOM)")
		return p.provisionWithExistingSBOM(ctx, ui, comm, generatedData)
	}

	// Native generation enabled
	ui.Say("Automatic SBOM generation enabled")
	p.warnDeprecatedConfigInUI(ui)

	osType, osArch, err := p.detectRemoteOS(ctx, ui, comm, generatedData)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to detect remote OS: %s", err))
		ui.Error("SBOM generation will be skipped, but build will continue")
		return nil
	}
	ui.Say(fmt.Sprintf("Detected: OS=%s, Arch=%s", osType, osArch))

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
	generatedData map[string]any,
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
	generatedData map[string]any) (string, string, error) {
	// First check if already detected (from generatedData)
	if osType, ok := generatedData["OSType"].(string); ok {
		if osArch, ok := generatedData["OSArch"].(string); ok {
			return osType, osArch, nil
		}
	}

	// Not in generatedData, detect now
	log.Println("Running OS detection commands on remote host...")

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
	log.Printf("OS detection output: %s", output)

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
			_ = tmpFile.Close() // Ignore error on close after getting name
		}
		return dst, nil
	}

	outDir := filepath.Dir(dst)
	// In case the destination does not exist, we'll get the dirpath,
	// and create it if it doesn't already exist
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create destination directory for user SBOM: %s", err)
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
		_ = tmpFile.Close() // Ignore error on close after getting name
	}

	return dst, nil
}

// uploadScanner uploads the scanner binary to the remote host.
// For Windows: uploads zip file and extracts remotely (optimization for slow WinRM uploads).
// For Unix: uploads binary directly.
//
// Unix path:
//  1. Extract local zip entry `packer`
//  2. Upload binary → /tmp/packer-sbom-runner
//  3. chmod +x      → /tmp/packer-sbom-runner
//
// Windows path:
//  1. Upload zip  → C:\Windows\Temp\packer-sbom-runner.zip
//     2-5. Single PowerShell command (Expand-Archive + Move-Item + Remove-Item)
func (p *Provisioner) uploadScanner(ctx context.Context, ui packersdk.Ui,
	comm packersdk.Communicator, localZipPath, osType string) (string, error) {
	_ = ui

	isWindows := strings.Contains(strings.ToLower(osType), "windows")

	var remotePath, binaryName string
	if isWindows {
		binaryName = "packer.exe"
		remotePath = "C:\\Windows\\Temp\\packer-sbom-runner.exe"
	} else {
		binaryName = "packer"
		remotePath = "/tmp/packer-sbom-runner"
	}

	if isWindows {
		remoteDir := "C:\\Windows\\Temp"
		remoteZipPath := remoteDir + "\\packer-sbom-runner.zip"

		// Step 1: upload zip to remote.
		zipFile, err := os.Open(localZipPath)
		if err != nil {
			return "", fmt.Errorf("failed to open Packer release zip: %s", err)
		}
		defer func() { _ = zipFile.Close() }()

		log.Printf("[INFO] Uploading Packer release zip to %s...", remoteZipPath)
		if err := comm.Upload(remoteZipPath, zipFile, nil); err != nil {
			return "", fmt.Errorf("failed to upload Packer release zip: %s", err)
		}

		// Single PowerShell command: extract, move binary, remove zip.
		psCmd := fmt.Sprintf(
			`powershell -NoProfile -ExecutionPolicy Bypass -Command `+
				`"$ErrorActionPreference='Stop'; `+
				`Expand-Archive -Path '%s' -DestinationPath '%s' -Force; `+
				`if (!(Test-Path '%s\%s')) { throw 'packer.exe not found after extraction' }; `+
				`Move-Item -Force '%s\%s' '%s'; `+
				`Remove-Item -Force '%s'"`,
			remoteZipPath, remoteDir,
			remoteDir, binaryName,
			remoteDir, binaryName, remotePath,
			remoteZipPath,
		)
		if err := p.runRemoteCmd(ctx, comm, psCmd, "extract scanner (Windows)"); err != nil {
			return "", err
		}
	} else {
		// Step 1: extract the binary locally from the release zip.
		binaryData, err := extractBinaryFromZip(localZipPath, binaryName)
		if err != nil {
			return "", fmt.Errorf("failed to extract %s from Packer release zip: %w", binaryName, err)
		}

		// Step 2: upload binary directly to remote.
		localFile := bytes.NewReader(binaryData)
		log.Printf("[INFO] Uploading Packer binary to %s...", remotePath)
		if err := comm.Upload(remotePath, localFile, nil); err != nil {
			return "", fmt.Errorf("failed to upload Packer binary: %s", err)
		}

		// Step 3: make it executable.
		chmodCmd := fmt.Sprintf(`chmod +x "%s"`, remotePath)
		if err := p.runRemoteCmd(ctx, comm, chmodCmd, "chmod scanner binary"); err != nil {
			return "", err
		}

		// Final verify: confirm binary is executable.
		verifyCmd := fmt.Sprintf(`test -x "%s"`, remotePath)
		if err := p.runRemoteCmd(ctx, comm, verifyCmd, "verify scanner is executable"); err != nil {
			return "", fmt.Errorf("scanner binary is not executable at %s after chmod; "+
				"check that /tmp is not mounted noexec on the remote host", remotePath)
		}
	}

	return remotePath, nil
}

func extractBinaryFromZip(zipPath, binaryName string) ([]byte, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer func() { _ = zr.Close() }()

	for _, f := range zr.File {
		if f.Name != binaryName {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open %s from zip: %w", binaryName, err)
		}

		data, readErr := io.ReadAll(rc)
		closeErr := rc.Close()
		if readErr != nil {
			return nil, fmt.Errorf("failed to read %s from zip: %w", binaryName, readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("failed to close %s stream from zip: %w", binaryName, closeErr)
		}
		return data, nil
	}

	return nil, fmt.Errorf("%s not found in zip %s", binaryName, zipPath)
}

// runRemoteCmd runs a single shell command on the remote host and returns a
// descriptive error if it fails.
func (p *Provisioner) runRemoteCmd(ctx context.Context, comm packersdk.Communicator, cmdStr, step string) error {
	var stderr bytes.Buffer
	cmd := &packersdk.RemoteCmd{
		Command: cmdStr,
		Stderr:  &stderr,
	}
	if err := comm.Start(ctx, cmd); err != nil {
		return fmt.Errorf("failed to start remote step %q: %s", step, err)
	}
	cmd.Wait()
	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("remote step %q failed (exit %d): %s",
			step, cmd.ExitStatus(), strings.TrimSpace(stderr.String()))
	}
	return nil
}

// provisionWithNativeGeneration handles the native SBOM generation flow.
// The Packer release binary is downloaded on the host (works in air-gapped
// environments where the VM has no internet access), verified, then uploaded
// to the remote host via the communicator before running `packer sbom-generate`.
func (p *Provisioner) provisionWithNativeGeneration(
	ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator,
	generatedData map[string]any, osType, osArch string,
) error {
	ui.Say("Starting Automatic SBOM generation workflow...")

	// Step 1: Download Packer release binary on the host.
	targetGOOS := strings.ToLower(osType)
	archMap := map[string]string{
		"x86_64": "amd64", "aarch64": "arm64", "i386": "386", "i686": "386", "armv7l": "arm", "armv7": "arm",
	}
	targetGOARCH := strings.ToLower(osArch)
	if mapped, ok := archMap[targetGOARCH]; ok {
		targetGOARCH = mapped
	}
	ui.Say(fmt.Sprintf("Downloading latest Packer release for %s/%s from %s...", targetGOOS, targetGOARCH, getReleaseBaseURL()))
	scannerZipPath, err := downloadPackerRelease(ctx, targetGOOS, targetGOARCH)
	if err != nil {
		return fmt.Errorf("failed to download Packer release for %s/%s: %w", targetGOOS, targetGOARCH, err)
	}
	defer func() {
		if err := os.Remove(scannerZipPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Printf("[WARN] failed to remove temporary Packer release zip %s: %v", scannerZipPath, err)
		}
	}()

	// Step 2: Upload scanner to remote.
	log.Println("Uploading scanner to remote host...")
	remoteScannerPath, err := p.uploadScanner(ctx, ui, comm, scannerZipPath, osType)
	if err != nil {
		return fmt.Errorf("failed to upload scanner: %s", err)
	}
	defer p.cleanupRemoteFile(ctx, ui, comm, remoteScannerPath)

	// Step 3: Run scanner on remote
	ui.Say(fmt.Sprintf("Running scanner on remote host (scanning %s)...", p.config.ScanPath))
	remoteSBOMPath, err := p.runScanner(ctx, ui, comm, remoteScannerPath, osType)
	if err != nil {
		return fmt.Errorf("failed to run scanner: %s", err)
	}
	defer p.cleanupRemoteFile(ctx, ui, comm, remoteSBOMPath)

	// Step 4: Download SBOM from remote
	log.Println("Downloading SBOM from remote host...")
	sbomData, err := p.downloadSBOM(ctx, ui, comm, remoteSBOMPath)
	if err != nil {
		return fmt.Errorf("failed to download SBOM: %s", err)
	}

	// Step 5: Process SBOM for HCP (validate, compress, store)
	log.Println("Processing SBOM for HCP Packer...")
	if err := p.processSBOMForHCP(generatedData, sbomData); err != nil {
		return fmt.Errorf("failed to process SBOM: %s", err)
	}

	ui.Say("Automatic SBOM generation completed successfully")
	return nil
}

// runScanner executes `packer sbom-generate` on the remote host.
func (p *Provisioner) runScanner(ctx context.Context, ui packersdk.Ui,
	comm packersdk.Communicator, scannerPath, osType string) (string, error) {

	// Determine output path based on OS
	var outputPath string
	isWindows := strings.Contains(strings.ToLower(osType), "windows")
	if isWindows {
		outputPath = "C:\\Windows\\Temp\\packer-sbom.json"
	} else {
		outputPath = "/tmp/packer-sbom.json"
	}

	// Restrict execute_command interpolation data to explicit scanner keys.
	// This avoids exposing arbitrary generatedData values to shell template rendering.
	templateData := map[string]string{
		"Path":     scannerPath,
		"Args":     strings.Join(p.config.ScannerArgs, " "),
		"ScanPath": p.config.ScanPath,
		"Output":   outputPath,
	}
	renderCtx := p.config.ctx
	renderCtx.Data = templateData

	// Use Windows-specific default if on Windows and user hasn't customized
	executeCommand := p.config.ExecuteCommand
	if isWindows && executeCommand == "chmod +x {{.Path}} && sudo {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}" {
		// User didn't customize, use Windows default (no sudo, uses sbom-generate subcommand).
		executeCommand = "{{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}"
	}

	// Backward compatibility: older execute_command templates omitted the
	// sbom-generate subcommand and invoked the scanner binary directly.
	normalizedExecuteCommand := normalizeScannerExecuteCommand(executeCommand)
	if normalizedExecuteCommand != executeCommand {
		log.Printf("[INFO] execute_command compatibility: injected 'sbom-generate' subcommand")
		executeCommand = normalizedExecuteCommand
	}

	// Render the execute command template
	cmdStr, err := interpolate.Render(executeCommand, &renderCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render execute_command: %s", err)
	}

	// For Windows with elevated user, wrap command with elevated runner
	if isWindows && p.config.ElevatedUser != "" {
		log.Printf("Using elevated user '%s' for scanner execution", p.config.ElevatedUser)
		elevatedCmd, err := guestexec.GenerateElevatedRunner(cmdStr, p)
		if err != nil {
			return "", fmt.Errorf("failed to generate elevated runner: %s", err)
		}
		cmdStr = elevatedCmd
	}

	log.Printf("Executing: %s", cmdStr)

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

func normalizeScannerExecuteCommand(executeCommand string) string {
	// Walk each {{.Path}} token and only inject "sbom-generate" when that
	// token is being used as the scanner executable invocation.
	//
	// Example rewritten:
	//   chmod +x {{.Path}} && {{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}
	// becomes:
	//   chmod +x {{.Path}} && {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}
	//
	// Example left unchanged:
	//   chmod +x {{.Path}} && {{.Path}} version
	// because the token after {{.Path}} is not {{.Args}} or {{.ScanPath}}.
	var out strings.Builder
	cursor := 0

	for {
		loc := scannerPathTokenRegexp.FindStringIndex(executeCommand[cursor:])
		if loc == nil {
			break
		}

		end := cursor + loc[1]
		out.WriteString(executeCommand[cursor:end])

		after := executeCommand[end:]
		trimmedAfter := strings.TrimLeft(after, " \t")

		if !hasSBOMGenerateSubcommandPrefix(trimmedAfter) && scannerArgsOrScanPathTokenPrefixRegexp.MatchString(trimmedAfter) {
			out.WriteString(" sbom-generate")
		}

		cursor = end
	}

	out.WriteString(executeCommand[cursor:])
	return out.String()
}

func hasSBOMGenerateSubcommandPrefix(s string) bool {
	// Treat sbom-generate as already present only when it is a complete shell
	// token prefix, not when it is part of a longer word.
	//
	// Matches:
	//   sbom-generate {{.Args}}
	//   sbom-generate; echo done
	//
	// Does not match:
	//   sbom-generate-custom
	const subcommand = "sbom-generate"
	if !strings.HasPrefix(s, subcommand) {
		return false
	}

	if len(s) == len(subcommand) {
		return true
	}

	next := s[len(subcommand)]
	switch next {
	case ' ', '\t', '\n', '\r', ';', '|', '&', '>', '<':
		return true
	default:
		return false
	}
}

// downloadSBOM downloads the SBOM file from the remote host
func (p *Provisioner) downloadSBOM(ctx context.Context, ui packersdk.Ui,
	comm packersdk.Communicator, remotePath string) ([]byte, error) {

	var buf bytes.Buffer
	log.Printf("Downloading SBOM from %s...", remotePath)

	if err := comm.Download(remotePath, &buf); err != nil {
		return nil, fmt.Errorf("failed to download SBOM: %s", err)
	}

	if buf.Len() == 0 {
		return nil, fmt.Errorf("downloaded SBOM is empty")
	}

	log.Printf("Downloaded SBOM (%d bytes)", buf.Len())
	return buf.Bytes(), nil
}

// cleanupRemoteFile removes a file from the remote host.
func (p *Provisioner) cleanupRemoteFile(ctx context.Context, ui packersdk.Ui,
	comm packersdk.Communicator, remotePath string) {

	if remotePath == "" {
		return
	}

	log.Printf("Cleaning up remote file: %s", remotePath)

	// Determine delete command based on path (Windows vs Unix)
	var cmdStr string
	if strings.Contains(remotePath, "C:\\") || strings.Contains(remotePath, "c:\\") {
		cmdStr = fmt.Sprintf("del /F /Q \"%s\"", remotePath)
	} else {
		cmdStr = fmt.Sprintf("rm -f \"%s\"", remotePath)
	}

	cmd := &packersdk.RemoteCmd{
		Command: cmdStr,
	}

	if err := comm.Start(ctx, cmd); err != nil {
		ui.Error(fmt.Sprintf("Failed to cleanup: %s", err))
		return
	}

	cmd.Wait()
	if cmd.ExitStatus() != 0 {
		ui.Error(fmt.Sprintf("Cleanup command failed for %s with exit status %d", remotePath, cmd.ExitStatus()))
	}
}

// processSBOMForHCP validates, compresses, and prepares SBOM for HCP upload
func (p *Provisioner) processSBOMForHCP(generatedData map[string]any, sbomData []byte) error {
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
	defer func() {
		_ = outFile.Close() // Cleanup, ignore error
	}()

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

// Communicator returns the communicator for elevated execution
func (p *Provisioner) Communicator() packersdk.Communicator {
	return p.communicator
}

// ElevatedUser returns the elevated user for Windows execution
func (p *Provisioner) ElevatedUser() string {
	return p.config.ElevatedUser
}

// ElevatedPassword returns the elevated password for Windows execution
func (p *Provisioner) ElevatedPassword() string {
	// Interpolate password if needed
	p.config.ctx.Data = p.generatedData
	elevatedPassword, _ := interpolate.Render(p.config.ElevatedPassword, &p.config.ctx)
	return elevatedPassword
}
