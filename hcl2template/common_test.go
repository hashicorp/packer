package hcl2template

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

func getBasicParser() *Parser {
	return &Parser{
		Parser: hclparse.NewParser(),
		BuilderSchemas: mapOfBuilder(map[string]packer.Builder{
			"amazon-ebs":     &MockBuilder{},
			"virtualbox-iso": &MockBuilder{},
		}),
		ProvisionersSchemas: mapOfProvisioner(map[string]packer.Provisioner{
			"shell": &MockProvisioner{},
			"file":  &MockProvisioner{},
		}),
		PostProcessorsSchemas: mapOfPostProcessor(map[string]packer.PostProcessor{
			"amazon-import": &MockPostProcessor{},
		}),
	}
}

type mapOfBuilder map[string]packer.Builder

func (mob mapOfBuilder) Has(builder string) bool {
	_, res := mob[builder]
	return res
}

func (mob mapOfBuilder) Start(builder string) (packer.Builder, error) {
	d, found := mob[builder]
	var err error
	if !found {
		err = fmt.Errorf("Unknown entry %s", builder)
	}
	return d, err
}

func (mob mapOfBuilder) List() []string {
	res := []string{}
	for k := range mob {
		res = append(res, k)
	}
	return res
}

type mapOfCommunicator map[string]packer.ConfigurableCommunicator

func (mob mapOfCommunicator) Start(communicator string) (packer.ConfigurableCommunicator, error) {
	c, found := mob[communicator]
	var err error
	if !found {
		err = fmt.Errorf("Unknown entry %s", communicator)
	}
	return c, err
}

type mapOfProvisioner map[string]packer.Provisioner

func (mop mapOfProvisioner) Has(provisioner string) bool {
	_, res := mop[provisioner]
	return res
}

func (mop mapOfProvisioner) Start(provisioner string) (packer.Provisioner, error) {
	p, found := mop[provisioner]
	var err error
	if !found {
		err = fmt.Errorf("Unknown provisioner %s", provisioner)
	}
	return p, err
}

func (mod mapOfProvisioner) List() []string {
	res := []string{}
	for k := range mod {
		res = append(res, k)
	}
	return res
}

type mapOfPostProcessor map[string]packer.PostProcessor

func (mop mapOfPostProcessor) Has(provisioner string) bool {
	_, res := mop[provisioner]
	return res
}

func (mop mapOfPostProcessor) Start(postProcessor string) (packer.PostProcessor, error) {
	p, found := mop[postProcessor]
	var err error
	if !found {
		err = fmt.Errorf("Unknown post-processor %s", postProcessor)
	}
	return p, err
}

func (mod mapOfPostProcessor) List() []string {
	res := []string{}
	for k := range mod {
		res = append(res, k)
	}
	return res
}

type parseTestArgs struct {
	filename string
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCfg, gotDiags := tt.parser.parse(tt.args.filename)
			if tt.parseWantDiags == (gotDiags == nil) {
				t.Fatalf("Parser.parse() unexpected diagnostics. %s", gotDiags)
			}
			if tt.parseWantDiagHasErrors != gotDiags.HasErrors() {
				t.Fatalf("Parser.parse() unexpected diagnostics HasErrors. %s", gotDiags)
			}
			if diff := cmp.Diff(tt.parseWantCfg, gotCfg,
				cmpopts.IgnoreUnexported(cty.Value{}, Source{}, ProvisionerBlock{}, PostProcessorBlock{}),
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
				cmpopts.IgnoreUnexported(packer.CoreBuild{},
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
