package packer

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/hashicorp/packer/version"
)

// A CoreBuild struct represents a single build job, the result of which should
// be a single machine image artifact. This artifact may be comprised of
// multiple files, of course, but it should be for only a single provider (such
// as VirtualBox, EC2, etc.).
type CoreBuild struct {
	BuildName          string
	Type               string
	Builder            packersdk.Builder
	BuilderConfig      interface{}
	BuilderType        string
	hooks              map[string][]packersdk.Hook
	Provisioners       []CoreBuildProvisioner
	PostProcessors     [][]CoreBuildPostProcessor
	CleanupProvisioner CoreBuildProvisioner
	TemplatePath       string
	Variables          map[string]string

	// Indicates whether the build is already initialized before calling Prepare(..)
	Prepared bool

	debug         bool
	force         bool
	onError       string
	l             sync.Mutex
	prepareCalled bool
}

// CoreBuildPostProcessor Keeps track of the post-processor and the
// configuration of the post-processor used within a build.
type CoreBuildPostProcessor struct {
	PostProcessor     packersdk.PostProcessor
	PType             string
	PName             string
	config            map[string]interface{}
	KeepInputArtifact *bool
}

// CoreBuildProvisioner keeps track of the provisioner and the configuration of
// the provisioner within the build.
type CoreBuildProvisioner struct {
	PType       string
	PName       string
	Provisioner packersdk.Provisioner
	config      []interface{}
}

// Returns the name of the build.
func (b *CoreBuild) Name() string {
	if b.BuildName != "" {
		return b.BuildName + "." + b.Type
	}
	return b.Type
}

// Prepare prepares the build by doing some initialization for the builder
// and any hooks. This _must_ be called prior to Run. The parameter is the
// overrides for the variables within the template (if any).
func (b *CoreBuild) Prepare() (warn []string, err error) {
	// For HCL2 templates, the builder and hooks are initialized when the
	// template is parsed. Calling Prepare(...) is not necessary
	if b.Prepared {
		b.prepareCalled = true
		return
	}

	b.l.Lock()
	defer b.l.Unlock()

	if b.prepareCalled {
		panic("prepare already called")
	}

	// Templates loaded from HCL2 will never get here. TODO: move this code into
	// a custom json area instead of just aborting early for HCL.
	b.prepareCalled = true

	packerConfig := map[string]interface{}{
		common.BuildNameConfigKey:     b.Type,
		common.BuilderTypeConfigKey:   b.BuilderType,
		common.CoreVersionConfigKey:   version.FormattedVersion(),
		common.DebugConfigKey:         b.debug,
		common.ForceConfigKey:         b.force,
		common.OnErrorConfigKey:       b.onError,
		common.TemplatePathKey:        b.TemplatePath,
		common.UserVariablesConfigKey: b.Variables,
	}

	// Prepare the builder
	generatedVars, warn, err := b.Builder.Prepare(b.BuilderConfig, packerConfig)
	if err != nil {
		log.Printf("Build '%s' prepare failure: %s\n", b.Type, err)
		return
	}

	// If the builder has provided a list of to-be-generated variables that
	// should be made accessible to provisioners, pass that list into
	// the provisioner prepare() so that the provisioner can appropriately
	// validate user input against what will become available.
	generatedPlaceholderMap := BasicPlaceholderData()
	if generatedVars != nil {
		for _, k := range generatedVars {
			generatedPlaceholderMap[k] = fmt.Sprintf("Build_%s. "+
				packerbuilderdata.PlaceholderMsg, k)
		}
	}

	// Prepare the provisioners
	for _, coreProv := range b.Provisioners {
		configs := make([]interface{}, len(coreProv.config), len(coreProv.config)+1)
		copy(configs, coreProv.config)
		configs = append(configs, packerConfig)
		configs = append(configs, generatedPlaceholderMap)

		if err = coreProv.Provisioner.Prepare(configs...); err != nil {
			return
		}
	}

	// Prepare the on-error-cleanup provisioner
	if b.CleanupProvisioner.PType != "" {
		configs := make([]interface{}, len(b.CleanupProvisioner.config), len(b.CleanupProvisioner.config)+1)
		copy(configs, b.CleanupProvisioner.config)
		configs = append(configs, packerConfig)
		configs = append(configs, generatedPlaceholderMap)
		err = b.CleanupProvisioner.Provisioner.Prepare(configs...)
		if err != nil {
			return
		}
	}

	// Prepare the post-processors
	for _, ppSeq := range b.PostProcessors {
		for _, corePP := range ppSeq {
			err = corePP.PostProcessor.Configure(corePP.config, packerConfig, generatedPlaceholderMap)
			if err != nil {
				return
			}
		}
	}

	return
}

// Runs the actual build. Prepare must be called prior to running this.
func (b *CoreBuild) Run(ctx context.Context, originalUi packersdk.Ui) ([]packersdk.Artifact, error) {
	if !b.prepareCalled {
		panic("Prepare must be called first")
	}

	// Copy the hooks
	hooks := make(map[string][]packersdk.Hook)
	for hookName, hookList := range b.hooks {
		hooks[hookName] = make([]packersdk.Hook, len(hookList))
		copy(hooks[hookName], hookList)
	}

	// Add a hook for the provisioners if we have provisioners
	if len(b.Provisioners) > 0 {
		hookedProvisioners := make([]*HookedProvisioner, len(b.Provisioners))
		for i, p := range b.Provisioners {
			var pConfig interface{}
			if len(p.config) > 0 {
				pConfig = p.config[0]
			}
			if b.debug {
				hookedProvisioners[i] = &HookedProvisioner{
					&DebuggedProvisioner{Provisioner: p.Provisioner},
					pConfig,
					p.PType,
				}
			} else {
				hookedProvisioners[i] = &HookedProvisioner{
					p.Provisioner,
					pConfig,
					p.PType,
				}
			}
		}

		if _, ok := hooks[packersdk.HookProvision]; !ok {
			hooks[packersdk.HookProvision] = make([]packersdk.Hook, 0, 1)
		}

		hooks[packersdk.HookProvision] = append(hooks[packersdk.HookProvision], &ProvisionHook{
			Provisioners: hookedProvisioners,
		})
	}

	if b.CleanupProvisioner.PType != "" {
		hookedCleanupProvisioner := &HookedProvisioner{
			b.CleanupProvisioner.Provisioner,
			b.CleanupProvisioner.config,
			b.CleanupProvisioner.PType,
		}
		hooks[packersdk.HookCleanupProvision] = []packersdk.Hook{&ProvisionHook{
			Provisioners: []*HookedProvisioner{hookedCleanupProvisioner},
		}}
	}

	hook := &packersdk.DispatchHook{Mapping: hooks}
	artifacts := make([]packersdk.Artifact, 0, 1)

	// The builder just has a normal Ui, but targeted
	builderUi := &TargetedUI{
		Target: b.Name(),
		Ui:     originalUi,
	}

	log.Printf("Running builder: %s", b.BuilderType)
	ts := CheckpointReporter.AddSpan(b.BuilderType, "builder", b.BuilderConfig)
	builderArtifact, err := b.Builder.Run(ctx, builderUi, hook)
	ts.End(err)
	if err != nil {
		return nil, err
	}

	// If there was no result, don't worry about running post-processors
	// because there is nothing they can do, just return.
	if builderArtifact == nil {
		return nil, nil
	}

	errors := make([]error, 0)
	keepOriginalArtifact := len(b.PostProcessors) == 0

	select {
	case <-ctx.Done():
		log.Println("Build was cancelled. Skipping post-processors.")
		return nil, nil
	default:
	}

	// Run the post-processors
PostProcessorRunSeqLoop:
	for _, ppSeq := range b.PostProcessors {
		priorArtifact := builderArtifact
		for i, corePP := range ppSeq {
			ppUi := &TargetedUI{
				Target: fmt.Sprintf("%s (%s)", b.Name(), corePP.PType),
				Ui:     originalUi,
			}

			if corePP.PName == corePP.PType {
				builderUi.Say(fmt.Sprintf("Running post-processor: %s", corePP.PType))
			} else {
				builderUi.Say(fmt.Sprintf("Running post-processor: %s (type %s)", corePP.PName, corePP.PType))
			}
			ts := CheckpointReporter.AddSpan(corePP.PType, "post-processor", corePP.config)
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
			if corePP.KeepInputArtifact != nil {
				if defaultKeep && *corePP.KeepInputArtifact == false && forceOverride {
					log.Printf("The %s post-processor forces "+
						"keep_input_artifact=true to preserve integrity of the"+
						"build chain. User-set keep_input_artifact=false will be"+
						"ignored.", corePP.PType)
				} else {
					// User overrides default.
					keep = *corePP.KeepInputArtifact
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
		err = &packersdk.MultiError{Errors: errors}
	}

	return artifacts, err
}

func (b *CoreBuild) SetDebug(val bool) {
	if b.prepareCalled {
		panic("prepare has already been called")
	}

	b.debug = val
}

func (b *CoreBuild) SetForce(val bool) {
	if b.prepareCalled {
		panic("prepare has already been called")
	}

	b.force = val
}

func (b *CoreBuild) SetOnError(val string) {
	if b.prepareCalled {
		panic("prepare has already been called")
	}

	b.onError = val
}
