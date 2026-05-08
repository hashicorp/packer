// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

//go:build !netbsd && !openbsd && !solaris && !mips && !mipsle && !mips64 && !(freebsd && 386)

package sbom

import (
	"context"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cataloging"
	"github.com/anchore/syft/syft/format"
	"github.com/anchore/syft/syft/format/cyclonedxjson"
	"github.com/anchore/syft/syft/format/spdxjson"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
)

// Generate creates an SBOM for the configured scan path and returns the encoded result.
func (g *Generator) Generate(ctx context.Context) ([]byte, error) {

	sourceInput := g.config.ScanPath
	src, err := syft.GetSource(ctx, sourceInput, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}
	defer src.Close()

	sbomCfg := syft.DefaultCreateSBOMConfig().
		WithSearchConfig(cataloging.SearchConfig{
			Scope: source.SquashedScope,
		}).
		WithParallelism(g.config.Parallelism)

	sbomResult, err := syft.CreateSBOM(ctx, src, sbomCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SBOM: %w", err)
	}

	return g.encodeToFormat(sbomResult)
}

// encodeToFormat encodes the SBOM to the requested format.
func (g *Generator) encodeToFormat(sbomData *sbom.SBOM) ([]byte, error) {
	switch g.config.Format {
	case FormatCycloneDX:
		encoder, err := cyclonedxjson.NewFormatEncoderWithConfig(
			cyclonedxjson.EncoderConfig{
				Version: "1.5",
				Pretty:  true,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create CycloneDX encoder: %w", err)
		}
		return format.Encode(*sbomData, encoder)

	case FormatSPDX:
		encoder, err := spdxjson.NewFormatEncoderWithConfig(
			spdxjson.EncoderConfig{
				Version: "2.3",
				Pretty:  true,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create SPDX encoder: %w", err)
		}
		return format.Encode(*sbomData, encoder)

	default:
		return nil, fmt.Errorf("unsupported format: %s (supported: cyclonedx, spdx)", g.config.Format)
	}
}
