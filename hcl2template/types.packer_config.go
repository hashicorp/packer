package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/template"
)

// PackerConfig represents a loaded packer config
type PackerConfig struct {
	Sources map[SourceRef]*Source

	Variables PackerV1Variables

	Builds Builds

	Communicators map[CommunicatorRef]*Communicator
}

type PackerV1Build struct {
	Builders       []*template.Builder
	Provisioners   []*template.Provisioner
	PostProcessors []*template.PostProcessor
}

func (pkrCfg *PackerConfig) ToV1Build() PackerV1Build {
	var diags hcl.Diagnostics
	res := PackerV1Build{}

	for _, build := range pkrCfg.Builds {
		communicator, _ := pkrCfg.Communicators[build.ProvisionerGroups.FirstCommunicatorRef()]

		for _, from := range build.Froms {
			source, found := pkrCfg.Sources[from.Src]
			if !found {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unknown " + sourceLabel + " reference",
					Detail:   "",
					Subject:  &from.HCL2Ref.DeclRange,
				})

				continue
			}

			// provisioners := build.ProvisionerGroups.FlatProvisioners()
			// postProcessors := build.PostProvisionerGroups.FlatProvisioners()

			_ = from
			_ = source
			_ = communicator
			// _ = provisioners
			// _ = postProcessors
		}
	}
	return res
}

func (pkrCfg *PackerConfig) ToTemplate() (*template.Template, error) {
	var result template.Template
	// var errs error

	result.Comments = nil
	result.Variables = pkrCfg.Variables.Variables()
	// TODO(azr): add sensitive variables

	builder := pkrCfg.ToV1Build()
	_ = builder

	// // Gather all the post-processors
	// if len(r.PostProcessors) > 0 {
	// 	result.PostProcessors = make([][]*PostProcessor, 0, len(r.PostProcessors))
	// }
	// for i, v := range r.PostProcessors {
	// 	// Parse the configurations. We need to do this because post-processors
	// 	// can take three different formats.
	// 	configs, err := r.parsePostProcessor(i, v)
	// 	if err != nil {
	// 		errs = multierror.Append(errs, err)
	// 		continue
	// 	}

	// 	// Parse the PostProcessors out of the configs
	// 	pps := make([]*PostProcessor, 0, len(configs))
	// 	for j, c := range configs {
	// 		var pp PostProcessor
	// 		if err := r.decoder(&pp, nil).Decode(c); err != nil {
	// 			errs = multierror.Append(errs, fmt.Errorf(
	// 				"post-processor %d.%d: %s", i+1, j+1, err))
	// 			continue
	// 		}

	// 		// Type is required
	// 		if pp.Type == "" {
	// 			errs = multierror.Append(errs, fmt.Errorf(
	// 				"post-processor %d.%d: type is required", i+1, j+1))
	// 			continue
	// 		}

	// 		// Set the raw configuration and delete any special keys
	// 		pp.Config = c

	// 		// The name defaults to the type if it isn't set
	// 		if pp.Name == "" {
	// 			pp.Name = pp.Type
	// 		}

	// 		delete(pp.Config, "except")
	// 		delete(pp.Config, "only")
	// 		delete(pp.Config, "keep_input_artifact")
	// 		delete(pp.Config, "type")
	// 		delete(pp.Config, "name")

	// 		if len(pp.Config) == 0 {
	// 			pp.Config = nil
	// 		}

	// 		pps = append(pps, &pp)
	// 	}

	// 	result.PostProcessors = append(result.PostProcessors, pps)
	// }

	// // Gather all the provisioners
	// if len(r.Provisioners) > 0 {
	// 	result.Provisioners = make([]*Provisioner, 0, len(r.Provisioners))
	// }
	// for i, v := range r.Provisioners {
	// 	var p Provisioner
	// 	if err := r.decoder(&p, nil).Decode(v); err != nil {
	// 		errs = multierror.Append(errs, fmt.Errorf(
	// 			"provisioner %d: %s", i+1, err))
	// 		continue
	// 	}

	// 	// Type is required before any richer validation
	// 	if p.Type == "" {
	// 		errs = multierror.Append(errs, fmt.Errorf(
	// 			"provisioner %d: missing 'type'", i+1))
	// 		continue
	// 	}

	// 	// Set the raw configuration and delete any special keys
	// 	p.Config = v.(map[string]interface{})

	// 	delete(p.Config, "except")
	// 	delete(p.Config, "only")
	// 	delete(p.Config, "override")
	// 	delete(p.Config, "pause_before")
	// 	delete(p.Config, "type")
	// 	delete(p.Config, "timeout")

	// 	if len(p.Config) == 0 {
	// 		p.Config = nil
	// 	}

	// 	result.Provisioners = append(result.Provisioners, &p)
	// }

	// // If we have errors, return those with a nil result
	// if errs != nil {
	// 	return nil, errs
	// }

	return &result, nil
}
