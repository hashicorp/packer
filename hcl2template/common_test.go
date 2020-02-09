package hcl2template

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	. "github.com/hashicorp/packer/hcl2template/internal"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

func getBasicParser() *Parser {
	return &Parser{
		Parser: hclparse.NewParser(),
		BuilderSchemas: packer.MapOfBuilder{
			"amazon-ebs":     func() (packer.Builder, error) { return &MockBuilder{}, nil },
			"virtualbox-iso": func() (packer.Builder, error) { return &MockBuilder{}, nil },
		},
		ProvisionersSchemas: packer.MapOfProvisioner{
			"shell": func() (packer.Provisioner, error) { return &MockProvisioner{}, nil },
			"file":  func() (packer.Provisioner, error) { return &MockProvisioner{}, nil },
		},
		PostProcessorsSchemas: packer.MapOfPostProcessor{
			"amazon-import": func() (packer.PostProcessor, error) { return &MockPostProcessor{}, nil },
		},
	}
}

type parseTestArgs struct {
	filename string
	vars     map[string]string
}

type parseTest struct {
	name   string
	parser *Parser
	args   parseTestArgs

	parseWantCfg           *PackerConfig
	parseWantDiags         bool
	parseWantDiagHasErrors bool

	getBuildsWantBuilds []packer.Build
	getBuildsWantDiags  bool
	// getBuildsWantDiagHasErrors bool
}

func testParse(t *testing.T, tests []parseTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCfg, gotDiags := tt.parser.parse(tt.args.filename, tt.args.vars)
			if tt.parseWantDiags == (gotDiags == nil) {
				t.Fatalf("Parser.parse() unexpected diagnostics. %s", gotDiags)
			}
			if tt.parseWantDiagHasErrors != gotDiags.HasErrors() {
				t.Fatalf("Parser.parse() unexpected diagnostics HasErrors. %s", gotDiags)
			}
			if diff := cmp.Diff(tt.parseWantCfg, gotCfg,
				cmpopts.IgnoreUnexported(
					cty.Value{},
					cty.Type{},
					Variable{},
					SourceBlock{},
					ProvisionerBlock{},
					PostProcessorBlock{},
				),
				cmpopts.IgnoreTypes(HCL2Ref{}),
				cmpopts.IgnoreTypes([]hcl.Range{}),
				cmpopts.IgnoreTypes(hcl.Range{}),
				cmpopts.IgnoreInterfaces(struct{ hcl.Expression }{}),
				cmpopts.IgnoreInterfaces(struct{ hcl.Body }{}),
			); diff != "" {
				t.Fatalf("Parser.parse() wrong packer config. %s", diff)
			}
			if gotDiags.HasErrors() {
				return
			}

			gotBuilds, gotDiags := tt.parser.getBuilds(gotCfg)
			if tt.getBuildsWantDiags == (gotDiags == nil) {
				t.Fatalf("Parser.getBuilds() unexpected diagnostics. %s", gotDiags)
			}
			if diff := cmp.Diff(tt.getBuildsWantBuilds, gotBuilds,
				cmpopts.IgnoreUnexported(
					cty.Value{},
					cty.Type{},
					packer.CoreBuild{},
					packer.CoreBuildProvisioner{},
					packer.CoreBuildPostProcessor{},
				),
			); diff != "" {
				t.Fatalf("Parser.getBuilds() wrong packer builds. %s", diff)
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
	}

	basicMockBuilder = &MockBuilder{
		Config: MockConfig{
			NestedMockConfig: basicNestedMockConfig,
			Nested:           basicNestedMockConfig,
			NestedSlice: []NestedMockConfig{
				basicNestedMockConfig,
				basicNestedMockConfig,
			},
		},
	}

	basicMockProvisioner = &MockProvisioner{
		Config: MockConfig{
			NotSquashed:      "value",
			NestedMockConfig: basicNestedMockConfig,
			Nested:           basicNestedMockConfig,
			NestedSlice: []NestedMockConfig{
				{},
			},
		},
	}
	basicMockPostProcessor = &MockPostProcessor{
		Config: MockConfig{
			NestedMockConfig: basicNestedMockConfig,
			Nested:           basicNestedMockConfig,
			NestedSlice: []NestedMockConfig{
				{},
			},
		},
	}
	basicMockCommunicator = &MockCommunicator{
		Config: MockConfig{
			NestedMockConfig: basicNestedMockConfig,
			Nested:           basicNestedMockConfig,
			NestedSlice: []NestedMockConfig{
				{},
			},
		},
	}
)
