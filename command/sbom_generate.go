package command

//Types to define SBOM generation command
// - Command struct (e.g., SBOMGenerateCommand)
// - Flags for the command (e.g., --format, --output)
// - Method to execute the command (e.g., Run)

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/packer/internal/sbom"
)

type SBOMGenerateCommand struct {
	Meta
}

func (cmd *SBOMGenerateCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(cmd.Ui)
	defer cleanup()

	cfg, ret := cmd.ParseArgs(args)
	if ret != 0 {
		return ret
	}
	return cmd.RunContext(ctx, cfg)
}

func (cmd *SBOMGenerateCommand) ParseArgs(args []string) (*sbom.Config, int) {
	cfg := &sbom.Config{
		ScanPath:    "/",
		Format:      sbom.FormatCycloneDX, // default format
		Parallelism: 4,                    // default parallelism
	}

	//Parse Syft Style args
	// Parse Syft-style arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-o", "--output":
			// Next arg is format
			if i+1 >= len(args) {
				cmd.Ui.Error("Missing value for -o flag")
				return cfg, 1
			}
			i++
			formatStr := args[i]

			// Parse format string
			format, err := sbom.ParseFormatFromArgs(formatStr)
			if err != nil {
				cmd.Ui.Error(err.Error())
				return cfg, 1
			}
			cfg.Format = format

		default:
			// Assume it's the scan path (positional argument)
			if !strings.HasPrefix(arg, "-") {
				cfg.ScanPath = arg
			}
		}
	}
	return cfg, 0
}

func (cmd *SBOMGenerateCommand) RunContext(ctx context.Context, cfg *sbom.Config) int {
	fmt.Fprintf(os.Stderr, "Generating %s SBOM for %s...\n", cfg.Format, cfg.ScanPath)

	// Create generator
	generator := sbom.NewGenerator(*cfg)

	// Generate SBOM
	sbomData, err := generator.Generate(ctx)
	if err != nil {
		cmd.Ui.Error(fmt.Sprintf("SBOM generation failed: %s", err))
		return 1
	}

	// Write to stdout (will be redirected to file via > operator)
	_, err = os.Stdout.Write(sbomData)
	if err != nil {
		cmd.Ui.Error(fmt.Sprintf("Failed to write SBOM: %s", err))
		return 1
	}

	fmt.Fprintln(os.Stderr, "✓ SBOM generation completed")

	return 0

}
func (c *SBOMGenerateCommand) Synopsis() string {
	return "Generate SBOM for the local system (internal use)"
}
func (c *SBOMGenerateCommand) Help() string {
	helpText := `
Usage: packer sbom-generate [options] <path>
  Generate a Software Bill of Materials (SBOM) for the local filesystem.
  This command is typically invoked internally by the hcp-sbom provisioner.
Options:
  -o <format>      Output format: cyclonedx-json, spdx-json (default: cyclonedx-json)
Arguments:
  <path>           Path to scan (default: /)
Examples:
  # Generate CycloneDX SBOM for root filesystem
  packer sbom-generate -o cyclonedx-json / > sbom.json
  # Generate SPDX SBOM
  packer sbom-generate -o spdx-json / > sbom.json
  # Scan specific directory
  packer sbom-generate -o cyclonedx-json /opt/app > app-sbom.json
Note: Output is written to stdout. Use shell redirection (>) to save to file.
`
	return strings.TrimSpace(helpText)
}
