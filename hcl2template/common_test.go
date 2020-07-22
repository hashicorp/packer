package hcl2template

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/packer/builder/null"
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
			"null":           func() (packer.Builder, error) { return &null.Builder{}, nil },
		},
		ProvisionersSchemas: packer.MapOfProvisioner{
			"shell": func() (packer.Provisioner, error) { return &MockProvisioner{}, nil },
			"file":  func() (packer.Provisioner, error) { return &MockProvisioner{}, nil },
		},
		PostProcessorsSchemas: packer.MapOfPostProcessor{
			"amazon-import": func() (packer.PostProcessor, error) { return &MockPostProcessor{}, nil },
			"manifest":      func() (packer.PostProcessor, error) { return &MockPostProcessor{}, nil },
		},
	}
}

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

	getBuildsWantBuilds []packer.Build
	getBuildsWantDiags  bool
	// getBuildsWantDiagHasErrors bool
}

func testParse(t *testing.T, tests []parseTest) {
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
			if diff := cmp.Diff(tt.parseWantCfg, gotCfg,
				cmpopts.IgnoreUnexported(
					PackerConfig{},
					cty.Value{},
					cty.Type{},
					Variable{},
					SourceBlock{},
					ProvisionerBlock{},
					PostProcessorBlock{},
				),
				cmpopts.IgnoreFields(PackerConfig{},
					"Cwd", // Cwd will change for every computer
				),
				cmpopts.IgnoreTypes(HCL2Ref{}),
				cmpopts.IgnoreTypes([]hcl.Range{}),
				cmpopts.IgnoreTypes(hcl.Range{}),
				cmpopts.IgnoreInterfaces(struct{ hcl.Expression }{}),
				cmpopts.IgnoreInterfaces(struct{ hcl.Body }{}),
			); diff != "" {
				t.Fatalf("Parser.parse() wrong packer config. %s", diff)
			}

			if gotCfg != nil && !tt.parseWantDiagHasErrors {
				gotInputVar := gotCfg.InputVariables
				for name, value := range tt.parseWantCfg.InputVariables {
					if variable, ok := gotInputVar[name]; ok {
						if diff := cmp.Diff(variable.DefaultValue.GoString(), value.DefaultValue.GoString()); diff != "" {
							t.Fatalf("Parser.parse(): unexpected default value for %s: %s", name, diff)
						}
						if diff := cmp.Diff(variable.VarfileValue.GoString(), value.VarfileValue.GoString()); diff != "" {
							t.Fatalf("Parser.parse(): varfile value differs for %s: %s", name, diff)
						}
					} else {
						t.Fatalf("Parser.parse() missing input variable. %s", name)
					}
				}

				gotLocalVar := gotCfg.LocalVariables
				for name, value := range tt.parseWantCfg.LocalVariables {
					if variable, ok := gotLocalVar[name]; ok {
						if variable.DefaultValue.GoString() != value.DefaultValue.GoString() {
							t.Fatalf("Parser.parse() local variable %s expected '%s' but was '%s'", name, value.DefaultValue.GoString(), variable.DefaultValue.GoString())
						}
					} else {
						t.Fatalf("Parser.parse() missing local variable. %s", name)
					}
				}
			}

			if gotDiags.HasErrors() {
				return
			}

			gotBuilds, gotDiags := gotCfg.GetBuilds(packer.GetBuildsOptions{})
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
					null.Builder{},
					HCL2Provisioner{},
					HCL2PostProcessor{},
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
		Tags: []MockTag{},
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
