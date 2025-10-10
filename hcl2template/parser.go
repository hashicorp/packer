// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/dynblock"
	"github.com/hashicorp/hcl/v2/hclparse"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/internal/dag"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

const (
	packerLabel            = "packer"
	sourceLabel            = "source"
	variablesLabel         = "variables"
	variableLabel          = "variable"
	localsLabel            = "locals"
	localLabel             = "local"
	dataSourceLabel        = "data"
	buildLabel             = "build"
	hcpPackerRegistryLabel = "hcp_packer_registry"
	communicatorLabel      = "communicator"
)

var configSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: packerLabel},
		{Type: sourceLabel, LabelNames: []string{"type", "name"}},
		{Type: variablesLabel},
		{Type: variableLabel, LabelNames: []string{"name"}},
		{Type: localsLabel},
		{Type: localLabel, LabelNames: []string{"name"}},
		{Type: dataSourceLabel, LabelNames: []string{"type", "name"}},
		{Type: buildLabel},
		{Type: hcpPackerRegistryLabel},
		{Type: communicatorLabel, LabelNames: []string{"type", "name"}},
	},
}

// packerBlockSchema is the schema for a top-level "packer" block in
// a configuration file.
var packerBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "required_version"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "required_plugins"},
	},
}

// Parser helps you parse HCL folders. It will parse an hcl file or directory
// and start builders, provisioners and post-processors to configure them with
// the parsed HCL and then return a []packersdk.Build. Packer will use that list
// of Builds to run everything in order.
type Parser struct {
	CorePackerVersion *version.Version

	CorePackerVersionString string

	PluginConfig *packer.PluginConfig

	ValidationOptions

	*hclparse.Parser
}

const (
	hcl2FileExt            = ".pkr.hcl"
	hcl2JsonFileExt        = ".pkr.json"
	hcl2VarFileExt         = ".pkrvars.hcl"
	hcl2VarJsonFileExt     = ".pkrvars.json"
	hcl2AutoVarFileExt     = ".auto.pkrvars.hcl"
	hcl2AutoVarJsonFileExt = ".auto.pkrvars.json"
)

// Parse will Parse all HCL files in filename. Path can be a folder or a file.
//
// Parse will first Parse packer and variables blocks, omitting the rest, which
// can be expanded with dynamic blocks. We need to evaluate all variables for
// that, so that data sources can expand dynamic blocks too.
//
// Parse returns a PackerConfig that contains configuration layout of a packer
// build; sources(builders)/provisioners/posts-processors will not be started
// and their contents won't be verified; Most syntax errors will cause an error,
// init should be called next to expand dynamic blocks and verify that used
// things do exist.
func (p *Parser) Parse(filename string, varFiles []string, argVars map[string]string) (*PackerConfig, hcl.Diagnostics) {
	var files []*hcl.File
	var diags hcl.Diagnostics

	// parse config files
	if filename != "" {
		hclFiles, jsonFiles, moreDiags := GetHCL2Files(filename, hcl2FileExt, hcl2JsonFileExt)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			// here this probably means that the file was not found, let's
			// simply leave early.
			return nil, diags
		}
		if len(hclFiles)+len(jsonFiles) == 0 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Could not find any config file in " + filename,
				Detail: "A config file must be suffixed with `.pkr.hcl` or " +
					"`.pkr.json`. A folder can be referenced.",
			})
		}
		for _, filename := range hclFiles {
			f, moreDiags := p.ParseHCLFile(filename)
			diags = append(diags, moreDiags...)
			files = append(files, f)
		}
		for _, filename := range jsonFiles {
			f, moreDiags := p.ParseJSONFile(filename)
			diags = append(diags, moreDiags...)
			files = append(files, f)
		}
		if diags.HasErrors() {
			return nil, diags
		}
	}

	basedir := filename
	if isDir, err := isDir(basedir); err == nil && !isDir {
		basedir = filepath.Dir(basedir)
	}
	wd, err := os.Getwd()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Could not find current working directory",
			Detail:   err.Error(),
		})
	}
	cfg := &PackerConfig{
		Basedir:                 basedir,
		Cwd:                     wd,
		CorePackerVersionString: p.CorePackerVersionString,
		HCPVars:                 map[string]cty.Value{},
		ValidationOptions:       p.ValidationOptions,
		parser:                  p,
		files:                   files,
	}

	for _, file := range files {
		coreVersionConstraints, moreDiags := sniffCoreVersionRequirements(file.Body)
		cfg.Packer.VersionConstraints = append(cfg.Packer.VersionConstraints, coreVersionConstraints...)
		diags = append(diags, moreDiags...)
	}

	// Before we go further, we'll check to make sure this version can read
	// all files, so we can produce a version-related error message rather than
	// potentially-confusing downstream errors.
	versionDiags := cfg.CheckCoreVersionRequirements(p.CorePackerVersion.Core())
	diags = append(diags, versionDiags...)
	if versionDiags.HasErrors() {
		return cfg, diags
	}

	// Looks for invalid arguments or unsupported block types
	{
		for _, file := range files {
			_, moreDiags := file.Body.Content(configSchema)
			diags = append(diags, moreDiags...)
		}
	}

	// Decode required_plugins blocks.
	//
	// Note: using `latest` ( or actually an empty string ) in a config file
	// does not work and packer will ask you to pick a version
	{
		for _, file := range files {
			diags = append(diags, cfg.decodeRequiredPluginsBlock(file)...)
		}
	}

	// Decode variable blocks so that they are available later on. Here locals
	// can use input variables so we decode input variables first.
	{
		for _, file := range files {
			diags = append(diags, cfg.decodeInputVariables(file)...)
		}

		for _, file := range files {
			morediags := p.decodeDatasources(file, cfg)
			diags = append(diags, morediags...)
		}

		for _, file := range files {
			moreLocals, morediags := parseLocalVariableBlocks(file)
			diags = append(diags, morediags...)
			cfg.LocalBlocks = append(cfg.LocalBlocks, moreLocals...)
		}

		diags = diags.Extend(cfg.checkForDuplicateLocalDefinition())
	}

	// parse var files
	{
		hclVarFiles, jsonVarFiles, moreDiags := GetHCL2Files(filename, hcl2AutoVarFileExt, hcl2AutoVarJsonFileExt)
		diags = append(diags, moreDiags...)

		// Combine all variable files into a single list, preserving the intended precedence and order.
		// The order is: auto-loaded HCL files, auto-loaded JSON files, followed by user-specified varFiles.
		// This ensures that user-specified files can override values from auto-loaded files,
		// and that their relative order is preserved exactly as specified by the user.
		variableFileNames := append(append(hclVarFiles, jsonVarFiles...), varFiles...)

		var variableFiles []*hcl.File

		for _, file := range variableFileNames {
			var (
				f         *hcl.File
				moreDiags hcl.Diagnostics
			)
			switch filepath.Ext(file) {
			case ".hcl":
				f, moreDiags = p.ParseHCLFile(file)
			case ".json":
				f, moreDiags = p.ParseJSONFile(file)
			default:
				moreDiags = hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Could not guess format of " + file,
						Detail:   "A var file must be suffixed with `.hcl` or `.json`.",
					},
				}
			}

			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			variableFiles = append(variableFiles, f)

		}

		diags = append(diags, cfg.collectInputVariableValues(os.Environ(), variableFiles, argVars)...)
	}

	return cfg, diags
}

// sniffCoreVersionRequirements does minimal parsing of the given body for
// "packer" blocks with "required_version" attributes, returning the
// requirements found.
//
// This is intended to maximize the chance that we'll be able to read the
// requirements (syntax errors notwithstanding) even if the config file contains
// constructs that might've been added in future versions
//
// This is a "best effort" sort of method which will return constraints it is
// able to find, but may return no constraints at all if the given body is
// so invalid that it cannot be decoded at all.
func sniffCoreVersionRequirements(body hcl.Body) ([]VersionConstraint, hcl.Diagnostics) {

	var sniffRootSchema = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: packerLabel,
			},
		},
	}

	rootContent, _, diags := body.PartialContent(sniffRootSchema)

	var constraints []VersionConstraint

	for _, block := range rootContent.Blocks {
		content, blockDiags := block.Body.Content(packerBlockSchema)
		diags = append(diags, blockDiags...)

		attr, exists := content.Attributes["required_version"]
		if !exists {
			continue
		}

		constraint, constraintDiags := decodeVersionConstraint(attr)
		diags = append(diags, constraintDiags...)
		if !constraintDiags.HasErrors() {
			constraints = append(constraints, constraint)
		}
	}

	return constraints, diags
}

func filterVarsFromLogs(inputOrLocal Variables) {
	for _, variable := range inputOrLocal {
		if !variable.Sensitive {
			continue
		}
		value := variable.Value()
		_ = cty.Walk(value, func(_ cty.Path, nested cty.Value) (bool, error) {
			if nested.IsWhollyKnown() && !nested.IsNull() && nested.Type().Equals(cty.String) {
				packersdk.LogSecretFilter.Set(nested.AsString())
			}
			return true, nil
		})
	}
}

func (cfg *PackerConfig) detectBuildPrereqDependencies() hcl.Diagnostics {
	var diags hcl.Diagnostics

	for _, ds := range cfg.Datasources {
		dependencies := GetVarsByType(ds.block, "data")
		dependencies = append(dependencies, GetVarsByType(ds.block, "local")...)

		for _, dep := range dependencies {
			// If something is locally aliased as `local` or `data`, we'll falsely
			// report it as a local variable, which is not necessarily what we
			// want to process here, so we continue.
			//
			// Note: this is kinda brittle, we should understand scopes to accurately
			// mark something from an expression as a reference to a local variable.
			// No real good solution for this now, besides maybe forbidding something
			// to be locally aliased as `local`.
			if len(dep) < 2 {
				continue
			}
			rs, err := NewRefStringFromDep(dep)
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to process datasource dependency",
					Detail: fmt.Sprintf("An error occurred while processing a dependency for data source %s: %s",
						ds.Name(), err),
				})
				continue
			}

			err = ds.RegisterDependency(rs)
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to register datasource dependency",
					Detail: fmt.Sprintf("An error occurred while registering %q as a dependency for data source %s: %s",
						rs, ds.Name(), err),
				})
			}
		}

		cfg.Datasources[ds.Ref()] = ds
	}

	for _, loc := range cfg.LocalBlocks {
		dependencies := FilterTraversalsByType(loc.Expr.Variables(), "data")
		dependencies = append(dependencies, FilterTraversalsByType(loc.Expr.Variables(), "local")...)

		for _, dep := range dependencies {
			// If something is locally aliased as `local` or `data`, we'll falsely
			// report it as a local variable, which is not necessarily what we
			// want to process here, so we continue.
			//
			// Note: this is kinda brittle, we should understand scopes to accurately
			// mark something from an expression as a reference to a local variable.
			// No real good solution for this now, besides maybe forbidding something
			// to be locally aliased as `local`.
			if len(dep) < 2 {
				continue
			}
			rs, err := NewRefStringFromDep(dep)
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to process local dependency",
					Detail: fmt.Sprintf("An error occurred while processing a dependency for local variable %s: %s",
						loc.LocalName, err),
				})
				continue
			}

			err = loc.RegisterDependency(rs)
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to register local dependency",
					Detail: fmt.Sprintf("An error occurred while registering %q as a dependency for local variable %s: %s",
						rs, loc.LocalName, err),
				})
			}
		}
	}

	return diags
}

func (cfg *PackerConfig) buildPrereqsDAG() (*dag.AcyclicGraph, error) {
	retGraph := dag.AcyclicGraph{}

	verticesMap := map[string]dag.Vertex{}

	var err error

	// Do a first pass to create all the vertices
	for ref := range cfg.Datasources {
		// We keep a reference to the datasource separately from where it
		// is used to avoid getting bit by the loop semantics.
		//
		// This `ds` local variable is the same object for every loop
		// so if we directly use the address of this object, we'll end
		// up referencing the last node of the loop for each vertex,
		// leading to implicit cycles.
		//
		// However by capturing it locally in this loop, we have a
		// reference to the actual datasource block, so it ends-up being
		// the right instance for each vertex.
		ds := cfg.Datasources[ref]
		v := retGraph.Add(&ds)
		verticesMap[fmt.Sprintf("data.%s", ds.Name())] = v
	}
	// Note: locals being references to the objects already, we can safely
	// use the reference returned by the local loop.
	for _, local := range cfg.LocalBlocks {
		v := retGraph.Add(local)
		verticesMap[fmt.Sprintf("local.%s", local.LocalName)] = v
	}

	// Connect the vertices together
	//
	// Vertices that don't have dependencies will be connected to the
	// root vertex of the graph
	for _, ds := range cfg.Datasources {
		dsName := fmt.Sprintf("data.%s", ds.Name())

		source := verticesMap[dsName]
		if source == nil {
			err = multierror.Append(err, fmt.Errorf("unable to find source vertex %q for dependency analysis, this is likely a Packer bug", dsName))
			continue
		}

		for _, dep := range ds.Dependencies {
			target := verticesMap[dep.String()]
			if target == nil {
				err = multierror.Append(err, fmt.Errorf("could not get dependency %q for %q, %q missing in template", dep.String(), dsName, dep.String()))
				continue
			}

			retGraph.Connect(dag.BasicEdge(source, target))
		}
	}
	for _, loc := range cfg.LocalBlocks {
		locName := fmt.Sprintf("local.%s", loc.LocalName)

		source := verticesMap[locName]
		if source == nil {
			err = multierror.Append(err, fmt.Errorf("unable to find source vertex %q for dependency analysis, this is likely a Packer bug", locName))
			continue
		}

		for _, dep := range loc.dependencies {
			target := verticesMap[dep.String()]

			if target == nil {
				err = multierror.Append(err, fmt.Errorf("could not get dependency %q for %q, %q missing in template", dep.String(), locName, dep.String()))
				continue
			}

			retGraph.Connect(dag.BasicEdge(source, target))
		}
	}

	if validateErr := retGraph.Validate(); validateErr != nil {
		err = multierror.Append(err, validateErr)
	}

	return &retGraph, err
}

func (cfg *PackerConfig) evaluateBuildPrereqs(skipDatasources bool) hcl.Diagnostics {
	diags := cfg.detectBuildPrereqDependencies()
	if diags.HasErrors() {
		return diags
	}

	graph, err := cfg.buildPrereqsDAG()
	if err != nil {
		return diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to prepare execution graph",
			Detail:   fmt.Sprintf("An error occurred while building the graph for datasources/locals: %s", err),
		})
	}

	walkFunc := func(v dag.Vertex) hcl.Diagnostics {
		var diags hcl.Diagnostics

		switch bl := v.(type) {
		case *DatasourceBlock:
			diags = cfg.evaluateDatasource(*bl, skipDatasources)
		case *LocalBlock:
			var val *Variable
			if cfg.LocalVariables == nil {
				cfg.LocalVariables = make(Variables)
			}
			val, diags = cfg.evaluateLocalVariable(bl)
			// Note: clumsy a bit, but we won't add the variable as `nil` here
			// unless no errors have been reported during evaluation.
			//
			// This prevents Packer from panicking down the line, as initialisation
			// doesn't stop if there are diags, so if `val` is nil, it crashes.
			if !diags.HasErrors() {
				cfg.LocalVariables[bl.LocalName] = val
			}
		default:
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "unsupported DAG node type",
				Detail: fmt.Sprintf("A node of type %q was added to the DAG, but cannot be "+
					"evaluated as it is unsupported. "+
					"This is a Packer bug, please report it so we can investigate.",
					reflect.TypeOf(v).String()),
			})
		}

		if diags.HasErrors() {
			return diags
		}

		return nil
	}

	for _, vtx := range graph.ReverseTopologicalOrder() {
		vtxDiags := walkFunc(vtx)
		if vtxDiags.HasErrors() {
			diags = diags.Extend(vtxDiags)
			return diags
		}
	}

	return nil
}

func (cfg *PackerConfig) Initialize(opts packer.InitializeOptions) hcl.Diagnostics {
	diags := cfg.InputVariables.ValidateValues()

	if opts.UseSequential {
		diags = diags.Extend(cfg.evaluateDatasources(opts.SkipDatasourcesExecution))
		diags = diags.Extend(cfg.evaluateLocalVariables(cfg.LocalBlocks))
	} else {
		diags = diags.Extend(cfg.evaluateBuildPrereqs(opts.SkipDatasourcesExecution))
	}

	filterVarsFromLogs(cfg.InputVariables)
	filterVarsFromLogs(cfg.LocalVariables)

	// parse the actual content // rest
	for _, file := range cfg.files {
		diags = append(diags, cfg.parser.parseConfig(file, cfg)...)
	}

	diags = append(diags, cfg.initializeBlocks()...)

	return diags
}

// parseConfig looks in the found blocks for everything that is not a variable
// block.
func (p *Parser) parseConfig(f *hcl.File, cfg *PackerConfig) hcl.Diagnostics {
	var diags hcl.Diagnostics

	body := f.Body
	body = dynblock.Expand(body, cfg.EvalContext(DatasourceContext, nil))
	content, moreDiags := body.Content(configSchema)
	diags = append(diags, moreDiags...)

	for _, block := range content.Blocks {
		switch block.Type {
		case buildHCPPackerRegistryLabel:
			if cfg.HCPPackerRegistry != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Only one " + buildHCPPackerRegistryLabel + " is allowed",
					Subject:  block.DefRange.Ptr(),
				})
				continue
			}
			hcpPackerRegistry, moreDiags := p.decodeHCPRegistry(block, cfg)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			cfg.HCPPackerRegistry = hcpPackerRegistry

		case sourceLabel:
			source, moreDiags := p.decodeSource(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			ref := source.Ref()
			if existing, found := cfg.Sources[ref]; found {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate " + sourceLabel + " block",
					Detail: fmt.Sprintf("This "+sourceLabel+" block has the "+
						"same builder type and name as a previous block declared "+
						"at %s. Each "+sourceLabel+" must have a unique name per builder type.",
						existing.block.DefRange.Ptr()),
					Subject: source.block.DefRange.Ptr(),
				})
				continue
			}

			if cfg.Sources == nil {
				cfg.Sources = map[SourceRef]SourceBlock{}
			}
			cfg.Sources[ref] = source

		case buildLabel:
			build, moreDiags := p.decodeBuildConfig(block, cfg)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			cfg.Builds = append(cfg.Builds, build)
		}
	}

	return diags
}

func (p *Parser) decodeDatasources(file *hcl.File, cfg *PackerConfig) hcl.Diagnostics {
	var diags hcl.Diagnostics

	body := file.Body
	content, _ := body.Content(configSchema)

	for _, block := range content.Blocks {
		switch block.Type {
		case dataSourceLabel:
			datasource, moreDiags := p.decodeDataBlock(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			ref := datasource.Ref()
			if existing, found := cfg.Datasources[ref]; found {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate " + dataSourceLabel + " block",
					Detail: fmt.Sprintf("This "+dataSourceLabel+" block has the "+
						"same data type and name as a previous block declared "+
						"at %s. Each "+dataSourceLabel+" must have a unique name per builder type.",
						existing.block.DefRange.Ptr()),
					Subject: datasource.block.DefRange.Ptr(),
				})
				continue
			}
			if cfg.Datasources == nil {
				cfg.Datasources = Datasources{}
			}
			cfg.Datasources[ref] = *datasource
		}
	}

	return diags
}
