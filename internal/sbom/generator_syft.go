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
	getSourceCfg := syft.DefaultGetSourceConfig()
	if len(g.config.Exclude) > 0 {
		getSourceCfg = getSourceCfg.WithExcludeConfig(source.ExcludeConfig{Paths: g.config.Exclude})
	}

	src, err := syft.GetSource(ctx, sourceInput, getSourceCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}
	defer src.Close()

	var scope source.Scope
	switch g.config.Scope {
	case ScopeAllLayers:
		scope = source.AllLayersScope
	case "", ScopeSquashed:
		scope = source.SquashedScope
	default:
		return nil, fmt.Errorf("unsupported scope: %s", g.config.Scope)
	}

	sbomCfg := syft.DefaultCreateSBOMConfig().
		WithSearchConfig(cataloging.SearchConfig{
			Scope: scope,
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
		cycloneCfg := cyclonedxjson.DefaultEncoderConfig()
		cycloneCfg.Pretty = true
		encoder, err := cyclonedxjson.NewFormatEncoderWithConfig(
			cycloneCfg,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create CycloneDX encoder: %w", err)
		}
		return format.Encode(*sbomData, encoder)

	case FormatSPDX:
		spdxCfg := spdxjson.DefaultEncoderConfig()
		spdxCfg.Pretty = true
		encoder, err := spdxjson.NewFormatEncoderWithConfig(
			spdxCfg,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create SPDX encoder: %w", err)
		}
		return format.Encode(*sbomData, encoder)

	default:
		return nil, fmt.Errorf("unsupported format: %s (supported: cyclonedx, spdx)", g.config.Format)
	}
}
