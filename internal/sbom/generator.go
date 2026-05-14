package sbom

import (
	"fmt"
	"strings"
)

// Format represents the SBOM output format
type Format string

const (
	FormatCycloneDX Format = "cyclonedx"
	FormatSPDX      Format = "spdx"
	ScopeSquashed   string = "squashed"
	ScopeAllLayers  string = "all-layers"
)

// Config holds configuration for SBOM generation
type Config struct {
	ScanPath    string // Path to scan (e.g., "/", "/opt/app")
	Format      Format // Output format (cyclonedx or spdx)
	Parallelism int    // Number of parallel catalogers (0 = auto-detect)
	Scope       string // Scan scope (squashed or all-layers)
	Exclude     []string
}

// Generator generates SBOMs using embedded Syft SDK
type Generator struct {
	config Config
}

// NewGenerator creates a new SBOM generator with the given configuration
func NewGenerator(cfg Config) *Generator {
	// Set defaults
	if cfg.ScanPath == "" {
		cfg.ScanPath = "/"
	}
	if cfg.Format == "" {
		cfg.Format = FormatCycloneDX
	}
	if cfg.Parallelism == 0 {
		cfg.Parallelism = 4
	}
	if cfg.Scope == "" {
		cfg.Scope = ScopeSquashed
	}

	return &Generator{config: cfg}
}

// ParseFormatFromArgs parses Syft-style format argument
// This is a PACKAGE-LEVEL EXPORTED FUNCTION
func ParseFormatFromArgs(formatArg string) (Format, error) {
	formatArg = strings.ToLower(formatArg)

	if strings.Contains(formatArg, "cyclonedx") {
		return FormatCycloneDX, nil
	}
	if strings.Contains(formatArg, "spdx") {
		return FormatSPDX, nil
	}

	return "", fmt.Errorf("unsupported format: %s", formatArg)
}

func ParseScopeFromArgs(scopeArg string) (string, error) {
	scopeArg = strings.ToLower(strings.TrimSpace(scopeArg))

	switch scopeArg {
	case ScopeSquashed:
		return ScopeSquashed, nil
	case ScopeAllLayers, "all", "alllayers":
		return ScopeAllLayers, nil
	default:
		return "", fmt.Errorf("unsupported scope: %s (supported: squashed, all-layers)", scopeArg)
	}
}
