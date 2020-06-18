package hcl2template

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

const (
	buildFromLabel = "from"

	buildSourceLabel = "source"

	buildProvisionerLabel = "provisioner"

	buildPostProcessorLabel = "post-processor"
)

var buildSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: buildFromLabel, LabelNames: []string{"type"}},
		{Type: sourceLabel, LabelNames: []string{"reference"}},
		{Type: buildProvisionerLabel, LabelNames: []string{"type"}},
		{Type: buildPostProcessorLabel, LabelNames: []string{"type"}},
	},
}

// BuildBlock references an HCL 'build' block and it content, for example :
//
//	build {
//		sources = [
//			...
//		]
//		provisioner "" { ... }
//		post-processor "" { ... }
//	}
type BuildBlock struct {
	// Name is a string representing the named build to show in the logs
	Name string

	// Sources is the list of sources that we want to start in this build block.
	Sources []SourceRef

	// ProvisionerBlocks references a list of HCL provisioner block that will
	// will be ran against the sources.
	ProvisionerBlocks []*ProvisionerBlock

	// ProvisionerBlocks references a list of HCL post-processors block that
	// will be ran against the artifacts from the provisioning steps.
	PostProcessors []*PostProcessorBlock

	HCL2Ref HCL2Ref
}

type Builds []*BuildBlock

// decodeBuildConfig is called when a 'build' block has been detected. It will
// load the references to the contents of the build block.
func (p *Parser) decodeBuildConfig(block *hcl.Block) (*BuildBlock, hcl.Diagnostics) {
	build := &BuildBlock{}

	var b struct {
		Name        string   `hcl:"name,optional"`
		FromSources []string `hcl:"sources,optional"`
		Config      hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)
	if diags.HasErrors() {
		return nil, diags
	}

	build.Name = b.Name

	for _, buildFrom := range b.FromSources {
		ref := sourceRefFromString(buildFrom)

		if ref == NoSource ||
			!hclsyntax.ValidIdentifier(ref.Type) ||
			!hclsyntax.ValidIdentifier(ref.Name) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid " + sourceLabel + " reference",
				Detail: "A " + sourceLabel + " type is made of three parts that are" +
					"split by a dot `.`; each part must start with a letter and " +
					"may contain only letters, digits, underscores, and dashes." +
					"A valid source reference looks like: `source.type.name`",
				Subject: block.DefRange.Ptr(),
			})
			continue
		}

		build.Sources = append(build.Sources, ref)
	}

	content, moreDiags := b.Config.Content(buildSchema)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return nil, diags
	}
	for _, block := range content.Blocks {
		switch block.Type {
		case sourceLabel:
			ref, moreDiags := p.decodeBuildSource(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			build.Sources = append(build.Sources, ref)
		case buildProvisionerLabel:
			p, moreDiags := p.decodeProvisioner(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			build.ProvisionerBlocks = append(build.ProvisionerBlocks, p)
		case buildPostProcessorLabel:
			pp, moreDiags := p.decodePostProcessor(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			build.PostProcessors = append(build.PostProcessors, pp)
		}
	}

	return build, diags
}

// A CoreHCL2Build struct represents a single build job, the result of which should
// be a single machine image artifact. This artifact may be comprised of
// multiple files, of course, but it should be for only a single provider (such
// as VirtualBox, EC2, etc.).
type CoreHCL2Build struct {
	BuildName          string
	Type               string
	Builder            packer.Builder
	BuilderConfig      interface{}
	BuilderType        string
	hooks              map[string][]packer.Hook
	Provisioners       []CoreHCL2BuildProvisioner
	PostProcessors     [][]CoreHCL2BuildPostProcessor
	CleanupProvisioner CoreHCL2BuildProvisioner
	TemplatePath       string
	Variables          map[string]string

	debug         bool
	force         bool
	onError       string
	l             sync.Mutex
	prepareCalled bool

	cfg *PackerConfig
}

// CoreHCL2BuildPostProcessor Keeps track of the post-processor and the
// configuration of the post-processor used within a build.
type CoreHCL2BuildPostProcessor struct {
	PostProcessor     packer.PostProcessor
	PType             string
	PName             string
	config            map[string]interface{}
	keepInputArtifact *bool
}

// CoreHCL2BuildProvisioner keeps track of the provisioner and the configuration of
// the provisioner within the build.
type CoreHCL2BuildProvisioner struct {
	PType       string
	PName       string
	Provisioner packer.Provisioner
	config      []interface{}
}

// Returns the name of the build.
func (b *CoreHCL2Build) Name() string {
	if b.BuildName != "" {
		return b.BuildName + "." + b.Type
	}
	return b.Type
}

// The build is prepared previously in PackerConfig.GetBuilds
// so don't do anything here.
func (b *CoreHCL2Build) Prepare() (warn []string, err error) {
	return nil, nil
}

// Runs the actual build. Prepare must be called prior to running this.
func (b *CoreHCL2Build) Run(ctx context.Context, originalUi packer.Ui) ([]packer.Artifact, error) {
	if !b.prepareCalled {
		panic("Prepare must be called first")
	}

	// Copy the hooks
	hooks := make(map[string][]packer.Hook)
	for hookName, hookList := range b.hooks {
		hooks[hookName] = make([]packer.Hook, len(hookList))
		copy(hooks[hookName], hookList)
	}

	// Add a hook for the provisioners if we have provisioners
	if len(b.Provisioners) > 0 {
		hookedProvisioners := make([]*packer.HookedProvisioner, len(b.Provisioners))
		for i, p := range b.Provisioners {
			var pConfig interface{}
			if len(p.config) > 0 {
				pConfig = p.config[0]
			}
			if b.debug {
				hookedProvisioners[i] = &packer.HookedProvisioner{
					&packer.DebuggedProvisioner{Provisioner: p.Provisioner},
					pConfig,
					p.PType,
				}
			} else {
				hookedProvisioners[i] = &packer.HookedProvisioner{
					p.Provisioner,
					pConfig,
					p.PType,
				}
			}
		}

		if _, ok := hooks[packer.HookProvision]; !ok {
			hooks[packer.HookProvision] = make([]packer.Hook, 0, 1)
		}

		hooks[packer.HookProvision] = append(hooks[packer.HookProvision], &packer.ProvisionHook{
			Provisioners: hookedProvisioners,
			HCL2Prepare: func(typeName string, data map[string]interface{}) (packer.Provisioner, hcl.Diagnostics) {
				// This will interpolate build variables by decoding the provisioner block again
				var diags hcl.Diagnostics
				if data == nil {
					diags = append(diags, &hcl.Diagnostic{
						Summary: fmt.Sprintf("failed loading %s", typeName),
						Detail:  "unable to prepare provisioner with build variables interpolation",
					})
					return nil, diags
				}

				for _, build := range b.cfg.Builds {
					for _, from := range build.Sources {
						src, _ := b.cfg.Sources[from.Ref()]

						variables := make(Variables)
						for k, v := range data {
							if value, ok := v.(string); ok {
								variables[k] = &Variable{
									DefaultValue: cty.StringVal(value),
									Type:         cty.String,
								}
							}
						}
						variablesVal, _ := variables.Values()

						generatedVariables := map[string]cty.Value{
							sourcesAccessor: cty.ObjectVal(map[string]cty.Value{
								"type": cty.StringVal(src.Type),
								"name": cty.StringVal(src.Name),
							}),
							buildAccessor: cty.ObjectVal(variablesVal),
						}

						for _, pb := range build.ProvisionerBlocks {
							if pb.PType != typeName {
								continue
							}
							return b.cfg.getProvisioner(src, pb, b.cfg.EvalContext(generatedVariables))
						}
					}
				}
				diags = append(diags, &hcl.Diagnostic{
					Summary: fmt.Sprintf("failed loading %s", typeName),
					Detail:  "unable to prepare provisioner with build variables interpolation",
				})
				return nil, diags
			},
		})
	}

	if b.CleanupProvisioner.PType != "" {
		hookedCleanupProvisioner := &packer.HookedProvisioner{
			b.CleanupProvisioner.Provisioner,
			b.CleanupProvisioner.config,
			b.CleanupProvisioner.PType,
		}
		hooks[packer.HookCleanupProvision] = []packer.Hook{&packer.ProvisionHook{
			Provisioners: []*packer.HookedProvisioner{hookedCleanupProvisioner},
		}}
	}

	hook := &packer.DispatchHook{Mapping: hooks}
	// The builder just has a normal Ui, but targeted
	builderUi := &packer.TargetedUI{
		Target: b.Name(),
		Ui:     originalUi,
	}

	log.Printf("Running builder: %s", b.BuilderType)
	ts := packer.CheckpointReporter.AddSpan(b.BuilderType, "builder", b.BuilderConfig)
	builderArtifact, err := b.Builder.Run(ctx, builderUi, hook)
	ts.End(err)
	if err != nil {
		return nil, err
	}

	artifacts := make([]packer.Artifact, 0, 1)

	// If there was no result, don't worry about running post-processors
	// because there is nothing they can do, just return.
	if builderArtifact == nil {
		return nil, nil
	}

	errors := make([]error, 0)
	keepOriginalArtifact := len(b.PostProcessors) == 0

	// This will interpolate build variables by decoding and preparing the post-processor block again
	generatedData := make(map[string]interface{})
	artifactSateData := builderArtifact.State("generated_data")
	if artifactSateData != nil {
		for k, v := range artifactSateData.(map[interface{}]interface{}) {
			generatedData[k.(string)] = v
		}
	}

	for _, build := range b.cfg.Builds {
		for _, from := range build.Sources {
			src, _ := b.cfg.Sources[from.Ref()]

			variables := make(Variables)
			for k, v := range generatedData {
				if value, ok := v.(string); ok {
					variables[k] = &Variable{
						DefaultValue: cty.StringVal(value),
						Type:         cty.String,
					}
				}
			}
			variablesVal, _ := variables.Values()

			generatedVariables := map[string]cty.Value{
				sourcesAccessor: cty.ObjectVal(map[string]cty.Value{
					"type": cty.StringVal(src.Type),
					"name": cty.StringVal(src.Name),
				}),
				buildAccessor: cty.ObjectVal(variablesVal),
			}

			postProcessors, diags := b.cfg.getCoreBuildPostProcessors(src, build.PostProcessors, b.cfg.EvalContext(generatedVariables))
			if diags.HasErrors() {
				errors = append(errors, diags)
			} else {
				b.PostProcessors = [][]CoreHCL2BuildPostProcessor{postProcessors}
			}
		}
	}

	// Run the post-processors
PostProcessorRunSeqLoop:
	for _, ppSeq := range b.PostProcessors {
		priorArtifact := builderArtifact
		for i, corePP := range ppSeq {

			ppUi := &packer.TargetedUI{
				Target: fmt.Sprintf("%s (%s)", b.Name(), corePP.PType),
				Ui:     originalUi,
			}

			if corePP.PName == corePP.PType {
				builderUi.Say(fmt.Sprintf("Running post-processor: %s", corePP.PType))
			} else {
				builderUi.Say(fmt.Sprintf("Running post-processor: %s (type %s)", corePP.PName, corePP.PType))
			}
			ts := packer.CheckpointReporter.AddSpan(corePP.PType, "post-processor", corePP.config)
			artifact, defaultKeep, forceOverride, err := corePP.PostProcessor.PostProcess(ctx, ppUi, priorArtifact)
			ts.End(err)
			if err != nil {
				errors = append(errors, fmt.Errorf("Post-processor failed: %s", err))
				continue PostProcessorRunSeqLoop
			}

			if artifact == nil {
				log.Println("Nil artifact, halting post-processor chain.")
				continue PostProcessorRunSeqLoop
			}

			keep := defaultKeep
			// When user has not set keep_input_artifact
			// corePP.keepInputArtifact is nil.
			// In this case, use the keepDefault provided by the postprocessor.
			// When user _has_ set keep_input_artifact, go with that instead.
			// Exception: for postprocessors that will fail/become
			// useless if keep isn't true, heed forceOverride and keep the
			// input artifact regardless of user preference.
			if corePP.keepInputArtifact != nil {
				if defaultKeep && *corePP.keepInputArtifact == false && forceOverride {
					log.Printf("The %s post-processor forces "+
						"keep_input_artifact=true to preserve integrity of the"+
						"build chain. User-set keep_input_artifact=false will be"+
						"ignored.", corePP.PType)
				} else {
					// User overrides default.
					keep = *corePP.keepInputArtifact
				}
			}
			if i == 0 {
				// This is the first post-processor. We handle deleting
				// previous artifacts a bit different because multiple
				// post-processors may be using the original and need it.
				if !keepOriginalArtifact && keep {
					log.Printf(
						"Flagging to keep original artifact from post-processor '%s'",
						corePP.PType)
					keepOriginalArtifact = true
				}
			} else {
				// We have a prior artifact. If we want to keep it, we append
				// it to the results list. Otherwise, we destroy it.
				if keep {
					artifacts = append(artifacts, priorArtifact)
				} else {
					log.Printf("Deleting prior artifact from post-processor '%s'", corePP.PType)
					if err := priorArtifact.Destroy(); err != nil {
						log.Printf("Error is %#v", err)
						errors = append(errors, fmt.Errorf("Failed cleaning up prior artifact: %s; pp is %s", err, corePP.PType))
					}
				}
			}

			priorArtifact = artifact
		}

		// Add on the last artifact to the results
		if priorArtifact != nil {
			artifacts = append(artifacts, priorArtifact)
		}
	}

	if keepOriginalArtifact {
		artifacts = append(artifacts, nil)
		copy(artifacts[1:], artifacts)
		artifacts[0] = builderArtifact
	} else {
		log.Printf("Deleting original artifact for build '%s'", b.Type)
		if err := builderArtifact.Destroy(); err != nil {
			errors = append(errors, fmt.Errorf("Error destroying builder artifact: %s; bad artifact: %#v", err, builderArtifact.Files()))
		}
	}

	if len(errors) > 0 {
		err = &packer.MultiError{errors}
	}

	return artifacts, err
}

func (b *CoreHCL2Build) SetDebug(val bool) {
	if b.prepareCalled {
		panic("prepare has already been called")
	}

	b.debug = val
}

func (b *CoreHCL2Build) SetForce(val bool) {
	if b.prepareCalled {
		panic("prepare has already been called")
	}

	b.force = val
}

func (b *CoreHCL2Build) SetOnError(val string) {
	if b.prepareCalled {
		panic("prepare has already been called")
	}

	b.onError = val
}
