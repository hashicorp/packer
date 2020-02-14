package packer

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/packer/helper/common"
)

func boolPointer(tf bool) *bool {
	return &tf
}

func testBuild() *CoreBuild {
	return &CoreBuild{
		Type:          "test",
		Builder:       &MockBuilder{ArtifactId: "b"},
		BuilderConfig: 42,
		BuilderType:   "foo",
		hooks: map[string][]Hook{
			"foo": {&MockHook{}},
		},
		Provisioners: []CoreBuildProvisioner{
			{
				PType:       "mock-provisioner",
				Provisioner: &MockProvisioner{},
				config:      []interface{}{42}},
		},
		PostProcessors: [][]CoreBuildPostProcessor{
			{
				{&MockPostProcessor{ArtifactId: "pp"}, "testPP", "testPPName", make(map[string]interface{}), boolPointer(true)},
			},
		},
		Variables: make(map[string]string),
		onError:   "cleanup",
	}
}

func testDefaultPackerConfig() map[string]interface{} {
	return map[string]interface{}{
		BuildNameConfigKey:     "test",
		BuilderTypeConfigKey:   "foo",
		DebugConfigKey:         false,
		ForceConfigKey:         false,
		OnErrorConfigKey:       "cleanup",
		TemplatePathKey:        "",
		UserVariablesConfigKey: make(map[string]string),
	}
}
func TestBuild_Name(t *testing.T) {
	build := testBuild()
	if build.Name() != "test" {
		t.Fatalf("bad: %s", build.Name())
	}
}

func TestBuild_Prepare(t *testing.T) {
	packerConfig := testDefaultPackerConfig()

	build := testBuild()
	builder := build.Builder.(*MockBuilder)

	build.Prepare()
	if !builder.PrepareCalled {
		t.Fatal("should be called")
	}
	if !reflect.DeepEqual(builder.PrepareConfig, []interface{}{42, packerConfig}) {
		t.Fatalf("bad: %#v", builder.PrepareConfig)
	}

	coreProv := build.Provisioners[0]
	prov := coreProv.Provisioner.(*MockProvisioner)
	if !prov.PrepCalled {
		t.Fatal("prep should be called")
	}
	if !reflect.DeepEqual(prov.PrepConfigs, []interface{}{42, packerConfig, BasicPlaceholderData()}) {
		t.Fatalf("bad: %#v", prov.PrepConfigs)
	}

	corePP := build.PostProcessors[0][0]
	pp := corePP.PostProcessor.(*MockPostProcessor)
	if !pp.ConfigureCalled {
		t.Fatal("should be called")
	}
	if !reflect.DeepEqual(pp.ConfigureConfigs, []interface{}{make(map[string]interface{}), packerConfig, BasicPlaceholderData()}) {
		t.Fatalf("bad: %#v", pp.ConfigureConfigs)
	}
}

func TestBuild_Prepare_SkipWhenBuilderAlreadyInitialized(t *testing.T) {
	build := testBuild()
	builder := build.Builder.(*MockBuilder)

	build.Prepared = true
	build.Prepare()
	if builder.PrepareCalled {
		t.Fatal("should not be called")
	}
}

func TestBuild_Prepare_Twice(t *testing.T) {
	build := testBuild()
	warn, err := build.Prepare()
	if len(warn) > 0 {
		t.Fatalf("bad: %#v", warn)
	}
	if err != nil {
		t.Fatalf("bad error: %s", err)
	}

	defer func() {
		p := recover()
		if p == nil {
			t.Fatalf("should've paniced")
		}

		if p.(string) != "prepare already called" {
			t.Fatalf("Invalid panic: %s", p)
		}
	}()

	build.Prepare()
}

func TestBuildPrepare_BuilderWarnings(t *testing.T) {
	expected := []string{"foo"}

	build := testBuild()
	builder := build.Builder.(*MockBuilder)
	builder.PrepareWarnings = expected

	warn, err := build.Prepare()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if !reflect.DeepEqual(warn, expected) {
		t.Fatalf("bad: %#v", warn)
	}
}

func TestBuild_Prepare_Debug(t *testing.T) {
	packerConfig := testDefaultPackerConfig()
	packerConfig[DebugConfigKey] = true

	build := testBuild()
	builder := build.Builder.(*MockBuilder)

	build.SetDebug(true)
	build.Prepare()
	if !builder.PrepareCalled {
		t.Fatalf("should be called")
	}
	if !reflect.DeepEqual(builder.PrepareConfig, []interface{}{42, packerConfig}) {
		t.Fatalf("bad: %#v", builder.PrepareConfig)
	}

	coreProv := build.Provisioners[0]
	prov := coreProv.Provisioner.(*MockProvisioner)
	if !prov.PrepCalled {
		t.Fatal("prepare should be called")
	}
	if !reflect.DeepEqual(prov.PrepConfigs, []interface{}{42, packerConfig, BasicPlaceholderData()}) {
		t.Fatalf("bad: %#v", prov.PrepConfigs)
	}
}

func TestBuildPrepare_variables_default(t *testing.T) {
	packerConfig := testDefaultPackerConfig()
	packerConfig[UserVariablesConfigKey] = map[string]string{
		"foo": "bar",
	}

	build := testBuild()
	build.Variables["foo"] = "bar"
	builder := build.Builder.(*MockBuilder)

	warn, err := build.Prepare()
	if len(warn) > 0 {
		t.Fatalf("bad: %#v", warn)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !builder.PrepareCalled {
		t.Fatal("prepare should be called")
	}

	if !reflect.DeepEqual(builder.PrepareConfig[1], packerConfig) {
		t.Fatalf("prepare bad: %#v", builder.PrepareConfig[1])
	}
}

func TestBuildPrepare_ProvisionerGetsGeneratedMap(t *testing.T) {
	packerConfig := testDefaultPackerConfig()

	build := testBuild()
	builder := build.Builder.(*MockBuilder)
	builder.GeneratedVars = []string{"PartyVar"}

	build.Prepare()
	if !builder.PrepareCalled {
		t.Fatalf("should be called")
	}
	if !reflect.DeepEqual(builder.PrepareConfig, []interface{}{42, packerConfig}) {
		t.Fatalf("bad: %#v", builder.PrepareConfig)
	}

	coreProv := build.Provisioners[0]
	prov := coreProv.Provisioner.(*MockProvisioner)
	if !prov.PrepCalled {
		t.Fatal("prepare should be called")
	}

	generated := BasicPlaceholderData()
	generated["PartyVar"] = "Build_PartyVar. " + common.PlaceholderMsg
	if !reflect.DeepEqual(prov.PrepConfigs, []interface{}{42, packerConfig, generated}) {
		t.Fatalf("bad: %#v", prov.PrepConfigs)
	}
}

func TestBuild_Run(t *testing.T) {
	ui := testUi()

	build := testBuild()
	build.Prepare()
	ctx := context.Background()
	artifacts, err := build.Run(ctx, ui)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("bad: %#v", artifacts)
	}

	// Verify builder was run
	builder := build.Builder.(*MockBuilder)
	if !builder.RunCalled {
		t.Fatal("should be called")
	}

	// Verify hooks are dispatchable
	dispatchHook := builder.RunHook
	dispatchHook.Run(ctx, "foo", nil, nil, 42)

	hook := build.hooks["foo"][0].(*MockHook)
	if !hook.RunCalled {
		t.Fatal("should be called")
	}
	if hook.RunData != 42 {
		t.Fatalf("bad: %#v", hook.RunData)
	}

	// Verify provisioners run
	dispatchHook.Run(ctx, HookProvision, nil, new(MockCommunicator), 42)
	prov := build.Provisioners[0].Provisioner.(*MockProvisioner)
	if !prov.ProvCalled {
		t.Fatal("should be called")
	}

	// Verify post-processor was run
	pp := build.PostProcessors[0][0].PostProcessor.(*MockPostProcessor)
	if !pp.PostProcessCalled {
		t.Fatal("should be called")
	}
}

func TestBuild_Run_Artifacts(t *testing.T) {
	ui := testUi()

	// Test case: Test that with no post-processors, we only get the
	// main build.
	build := testBuild()
	build.PostProcessors = [][]CoreBuildPostProcessor{}

	build.Prepare()
	artifacts, err := build.Run(context.Background(), ui)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds := []string{"b"}
	artifactIds := make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}

	// Test case: Test that with a single post-processor that doesn't keep
	// inputs, only that post-processors results are returned.
	build = testBuild()
	build.PostProcessors = [][]CoreBuildPostProcessor{
		{
			{&MockPostProcessor{ArtifactId: "pp"}, "pp", "testPPName", make(map[string]interface{}), boolPointer(false)},
		},
	}

	build.Prepare()
	artifacts, err = build.Run(context.Background(), ui)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds = []string{"pp"}
	artifactIds = make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}

	// Test case: Test that with multiple post-processors, as long as one
	// keeps the original, the original is kept.
	build = testBuild()
	build.PostProcessors = [][]CoreBuildPostProcessor{
		{
			{&MockPostProcessor{ArtifactId: "pp1"}, "pp", "testPPName", make(map[string]interface{}), boolPointer(false)},
		},
		{
			{&MockPostProcessor{ArtifactId: "pp2"}, "pp", "testPPName", make(map[string]interface{}), boolPointer(true)},
		},
	}

	build.Prepare()
	artifacts, err = build.Run(context.Background(), ui)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds = []string{"b", "pp1", "pp2"}
	artifactIds = make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}

	// Test case: Test that with sequences, intermediaries are kept if they
	// want to be.
	build = testBuild()
	build.PostProcessors = [][]CoreBuildPostProcessor{
		{
			{&MockPostProcessor{ArtifactId: "pp1a"}, "pp", "testPPName", make(map[string]interface{}), boolPointer(false)},
			{&MockPostProcessor{ArtifactId: "pp1b"}, "pp", "testPPName", make(map[string]interface{}), boolPointer(true)},
		},
		{
			{&MockPostProcessor{ArtifactId: "pp2a"}, "pp", "testPPName", make(map[string]interface{}), boolPointer(false)},
			{&MockPostProcessor{ArtifactId: "pp2b"}, "pp", "testPPName", make(map[string]interface{}), boolPointer(false)},
		},
	}

	build.Prepare()
	artifacts, err = build.Run(context.Background(), ui)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds = []string{"pp1a", "pp1b", "pp2b"}
	artifactIds = make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}

	// Test case: Test that with a single post-processor that forcibly
	// keeps inputs, that the artifacts are kept.
	build = testBuild()
	build.PostProcessors = [][]CoreBuildPostProcessor{
		{
			{
				&MockPostProcessor{ArtifactId: "pp", Keep: true, ForceOverride: true}, "pp", "testPPName", make(map[string]interface{}), boolPointer(false),
			},
		},
	}

	build.Prepare()

	artifacts, err = build.Run(context.Background(), ui)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds = []string{"b", "pp"}
	artifactIds = make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}

	// Test case: Test that with a single post-processor that non-forcibly
	// keeps inputs, that the artifacts are discarded if user overrides.
	build = testBuild()
	build.PostProcessors = [][]CoreBuildPostProcessor{
		{
			{
				&MockPostProcessor{ArtifactId: "pp", Keep: true, ForceOverride: false}, "pp", "testPPName", make(map[string]interface{}), boolPointer(false),
			},
		},
	}

	build.Prepare()
	artifacts, err = build.Run(context.Background(), ui)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds = []string{"pp"}
	artifactIds = make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}

	// Test case: Test that with a single post-processor that non-forcibly
	// keeps inputs, that the artifacts are kept if user does not have preference.
	build = testBuild()
	build.PostProcessors = [][]CoreBuildPostProcessor{
		{
			{
				&MockPostProcessor{ArtifactId: "pp", Keep: true, ForceOverride: false}, "pp", "testPPName", make(map[string]interface{}), nil,
			},
		},
	}

	build.Prepare()
	artifacts, err = build.Run(context.Background(), ui)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds = []string{"b", "pp"}
	artifactIds = make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}
}

func TestBuild_RunBeforePrepare(t *testing.T) {
	defer func() {
		p := recover()
		if p == nil {
			t.Fatal("should panic")
		}

		if p.(string) != "Prepare must be called first" {
			t.Fatalf("bad: %s", p.(string))
		}
	}()

	testBuild().Run(context.Background(), testUi())
}

func TestBuild_Cancel(t *testing.T) {
	build := testBuild()

	build.Prepare()

	topCtx, topCtxCancel := context.WithCancel(context.Background())

	builder := build.Builder.(*MockBuilder)

	builder.RunFn = func(ctx context.Context) {
		topCtxCancel()
	}

	_, err := build.Run(topCtx, testUi())
	if err == nil {
		t.Fatal("build should err")
	}
}
