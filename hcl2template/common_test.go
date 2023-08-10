// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/builder/null"
	dnull "github.com/hashicorp/packer/datasource/null"
	. "github.com/hashicorp/packer/hcl2template/internal"
	hcl2template "github.com/hashicorp/packer/hcl2template/internal"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

const lockedVersion = "v1.5.0"

func getBasicParser(opts ...getParserOption) *Parser {
	parser := &Parser{
		CorePackerVersion:       version.Must(version.NewSemver(lockedVersion)),
		CorePackerVersionString: lockedVersion,
		Parser:                  hclparse.NewParser(),
		PluginConfig: &packer.PluginConfig{
			Builders: packer.MapOfBuilder{
				"amazon-ebs":     func() (packersdk.Builder, error) { return &MockBuilder{}, nil },
				"virtualbox-iso": func() (packersdk.Builder, error) { return &MockBuilder{}, nil },
				"null":           func() (packersdk.Builder, error) { return &null.Builder{}, nil },
			},
			Provisioners: packer.MapOfProvisioner{
				"shell": func() (packersdk.Provisioner, error) { return &MockProvisioner{}, nil },
				"file":  func() (packersdk.Provisioner, error) { return &MockProvisioner{}, nil },
			},
			PostProcessors: packer.MapOfPostProcessor{
				"amazon-import": func() (packersdk.PostProcessor, error) { return &MockPostProcessor{}, nil },
				"manifest":      func() (packersdk.PostProcessor, error) { return &MockPostProcessor{}, nil },
			},
			DataSources: packer.MapOfDatasource{
				"amazon-ami": func() (packersdk.Datasource, error) { return &MockDatasource{}, nil },
				"null":       func() (packersdk.Datasource, error) { return &dnull.Datasource{}, nil },
			},
		},
	}
	for _, configure := range opts {
		configure(parser)
	}
	return parser
}

type getParserOption func(*Parser)

type parseTestArgs struct {
	filename string
	vars     map[string]string
	varFiles []string
}

type parseTest struct {
	name   string
	parser *Parser
	args   parseTestArgs

	parseWantCfg           *PackerConfig
	parseWantDiags         bool
	parseWantDiagHasErrors bool

	getBuildsWantBuilds []packersdk.Build
	getBuildsWantDiags  bool
	// getBuildsWantDiagHasErrors bool
}

func testParse(t *testing.T, tests []parseTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCfg, gotDiags := tt.parser.Parse(tt.args.filename, tt.args.varFiles, tt.args.vars)
			moreDiags := gotCfg.Initialize(packer.InitializeOptions{})
			gotDiags = append(gotDiags, moreDiags...)
			if tt.parseWantDiags == (gotDiags == nil) {
				t.Fatalf("Parser.parse() unexpected %q diagnostics.", gotDiags)
			}
			if tt.parseWantDiagHasErrors != gotDiags.HasErrors() {
				t.Fatalf("Parser.parse() unexpected diagnostics HasErrors. %s", gotDiags)
			}
			if diff := cmp.Diff(tt.parseWantCfg, gotCfg, cmpOpts...); diff != "" {
				t.Fatalf("Parser.parse() wrong packer config. %s", diff)
			}

			if gotCfg != nil && !tt.parseWantDiagHasErrors {
				if diff := cmp.Diff(tt.parseWantCfg.InputVariables, gotCfg.InputVariables, cmpOpts...); diff != "" {
					t.Fatalf("Parser.parse() unexpected input vars. %s", diff)
				}

				if diff := cmp.Diff(tt.parseWantCfg.LocalVariables, gotCfg.LocalVariables, cmpOpts...); diff != "" {
					t.Fatalf("Parser.parse() unexpected local vars. %s", diff)
				}
			}

			if gotDiags.HasErrors() {
				return
			}

			gotBuilds, gotDiags := gotCfg.GetBuilds(packer.GetBuildsOptions{})
			if tt.getBuildsWantDiags == (gotDiags == nil) {
				t.Fatalf("Parser.getBuilds() unexpected diagnostics. %s", gotDiags)
			}
			if diff := cmp.Diff(tt.getBuildsWantBuilds, gotBuilds, cmpOpts...); diff != "" {
				t.Fatalf("Parser.getBuilds() wrong packer builds. %s", diff)
			}
		})
	}
}

func testParse_only_Parse(t *testing.T, tests []parseTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCfg, gotDiags := tt.parser.Parse(tt.args.filename, tt.args.varFiles, tt.args.vars)
			if tt.parseWantDiags == (gotDiags == nil) {
				t.Fatalf("Parser.parse() unexpected %q diagnostics.", gotDiags)
			}
			if tt.parseWantDiagHasErrors != gotDiags.HasErrors() {
				t.Fatalf("Parser.parse() unexpected diagnostics HasErrors. %s", gotDiags)
			}
			if diff := cmp.Diff(tt.parseWantCfg, gotCfg, cmpOpts...); diff != "" {
				t.Fatalf("Parser.parse() wrong packer config. %s", diff)
			}

			if gotCfg != nil && !tt.parseWantDiagHasErrors {
				if diff := cmp.Diff(tt.parseWantCfg.InputVariables, gotCfg.InputVariables, cmpOpts...); diff != "" {
					t.Fatalf("Parser.parse() unexpected input vars. %s", diff)
				}

				if diff := cmp.Diff(tt.parseWantCfg.LocalVariables, gotCfg.LocalVariables, cmpOpts...); diff != "" {
					t.Fatalf("Parser.parse() unexpected local vars. %s", diff)
				}
			}

			if gotDiags.HasErrors() {
				return
			}
		})
	}
}

var (
	// everything in the tests is a basicNestedMockConfig this allow to test
	// each known type to packer ( and embedding ) in one go.
	basicNestedMockConfig = NestedMockConfig{
		String:   "string",
		Int:      42,
		Int64:    43,
		Bool:     true,
		Trilean:  config.TriTrue,
		Duration: 10 * time.Second,
		MapStringString: map[string]string{
			"a": "b",
			"c": "d",
		},
		SliceString: []string{
			"a",
			"b",
			"c",
		},
		SliceSliceString: [][]string{
			{"a", "b"},
			{"c", "d"},
		},
		Tags: []MockTag{},
	}

	// everything in the tests is a basicNestedMockConfig this allow to test
	// each known type to packer ( and embedding ) in one go.
	builderBasicNestedMockConfig = NestedMockConfig{
		String:   "string",
		Int:      42,
		Int64:    43,
		Bool:     true,
		Trilean:  config.TriTrue,
		Duration: 10 * time.Second,
		MapStringString: map[string]string{
			"a": "b",
			"c": "d",
		},
		SliceString: []string{
			"a",
			"b",
			"c",
		},
		SliceSliceString: [][]string{
			{"a", "b"},
			{"c", "d"},
		},
		Tags:       []MockTag{},
		Datasource: "string",
	}

	basicMockBuilder = &MockBuilder{
		Config: MockConfig{
			NestedMockConfig: builderBasicNestedMockConfig,
			Nested:           builderBasicNestedMockConfig,
			NestedSlice: []NestedMockConfig{
				builderBasicNestedMockConfig,
				builderBasicNestedMockConfig,
			},
		},
	}

	basicMockProvisioner = &MockProvisioner{
		Config: MockConfig{
			NotSquashed:      "value <UNKNOWN>",
			NestedMockConfig: basicNestedMockConfig,
			Nested:           basicNestedMockConfig,
			NestedSlice: []NestedMockConfig{
				{
					Tags: dynamicTagList,
				},
			},
		},
	}
	basicMockPostProcessor = &MockPostProcessor{
		Config: MockConfig{
			NotSquashed:      "value <UNKNOWN>",
			NestedMockConfig: basicNestedMockConfig,
			Nested:           basicNestedMockConfig,
			NestedSlice: []NestedMockConfig{
				{
					Tags: []MockTag{},
				},
			},
		},
	}
	basicMockPostProcessorDynamicTags = &MockPostProcessor{
		Config: MockConfig{
			NotSquashed: "value <UNKNOWN>",
			NestedMockConfig: NestedMockConfig{
				String:   "string",
				Int:      42,
				Int64:    43,
				Bool:     true,
				Trilean:  config.TriTrue,
				Duration: 10 * time.Second,
				MapStringString: map[string]string{
					"a": "b",
					"c": "d",
				},
				SliceString: []string{
					"a",
					"b",
					"c",
				},
				SliceSliceString: [][]string{
					{"a", "b"},
					{"c", "d"},
				},
				Tags: []MockTag{
					{Key: "first_tag_key", Value: "first_tag_value"},
					{Key: "Component", Value: "user-service"},
					{Key: "Environment", Value: "production"},
				},
			},
			Nested:      basicNestedMockConfig,
			NestedSlice: []NestedMockConfig{},
		},
	}
	basicMockCommunicator = &MockCommunicator{
		Config: MockConfig{
			NestedMockConfig: basicNestedMockConfig,
			Nested:           basicNestedMockConfig,
			NestedSlice: []NestedMockConfig{
				{
					Tags: []MockTag{},
				},
			},
		},
	}

	emptyMockBuilder = &MockBuilder{
		Config: MockConfig{
			NestedMockConfig: NestedMockConfig{
				Tags: []MockTag{},
			},
			Nested:      NestedMockConfig{},
			NestedSlice: []NestedMockConfig{},
		},
	}

	emptyMockProvisioner = &MockProvisioner{
		Config: MockConfig{
			NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
			NestedSlice:      []NestedMockConfig{},
		},
	}

	dynamicTagList = []MockTag{
		{
			Key:   "first_tag_key",
			Value: "first_tag_value",
		},
		{
			Key:   "Component",
			Value: "user-service",
		},
		{
			Key:   "Environment",
			Value: "production",
		},
	}
)

var ctyValueComparer = cmp.Comparer(func(x, y cty.Value) bool {
	return x.RawEquals(y)
})

var ctyTypeComparer = cmp.Comparer(func(x, y cty.Type) bool {
	if x == cty.NilType && y == cty.NilType {
		return true
	}
	if x == cty.NilType || y == cty.NilType {
		return false
	}
	return x.Equals(y)
})

var versionComparer = cmp.Comparer(func(x, y *version.Version) bool {
	return x.Equal(y)
})

var versionConstraintComparer = cmp.Comparer(func(x, y *version.Constraint) bool {
	return x.String() == y.String()
})

var cmpOpts = []cmp.Option{
	ctyValueComparer,
	ctyTypeComparer,
	versionComparer,
	versionConstraintComparer,
	cmpopts.IgnoreUnexported(
		PackerConfig{},
		Variable{},
		SourceBlock{},
		DatasourceBlock{},
		ProvisionerBlock{},
		PostProcessorBlock{},
		packer.CoreBuild{},
		HCL2Provisioner{},
		HCL2PostProcessor{},
		packer.CoreBuildPostProcessor{},
		packer.CoreBuildProvisioner{},
		packer.CoreBuildPostProcessor{},
		null.Builder{},
	),
	cmpopts.IgnoreFields(PackerConfig{},
		"Cwd",     // Cwd will change for every os type
		"HCPVars", // HCPVars will not be filled-in during parsing
	),
	cmpopts.IgnoreFields(VariableAssignment{},
		"Expr", // its an interface
	),
	cmpopts.IgnoreFields(packer.CoreBuild{},
		"HCLConfig",
	),
	cmpopts.IgnoreFields(packer.CoreBuildProvisioner{},
		"HCLConfig",
	),
	cmpopts.IgnoreFields(packer.CoreBuildPostProcessor{},
		"HCLConfig",
	),
	cmpopts.IgnoreTypes(hcl2template.MockBuilder{}),
	cmpopts.IgnoreTypes(HCL2Ref{}),
	cmpopts.IgnoreTypes([]*LocalBlock{}),
	cmpopts.IgnoreTypes([]hcl.Range{}),
	cmpopts.IgnoreTypes(hcl.Range{}),
	cmpopts.IgnoreInterfaces(struct{ hcl.Expression }{}),
	cmpopts.IgnoreInterfaces(struct{ hcl.Body }{}),
}
