package schedulers

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/hcl2template"
)

func (s *HCLSequentialScheduler) PrepareBuilds() hcl.Diagnostics {
	// verify that all used plugins do exist
	var diags hcl.Diagnostics

	if len(s.config.Builds) == 0 {
		return diags.Append(&hcl.Diagnostic{
			Summary:  "Missing build block",
			Detail:   "A build block with one or more sources is required for executing a build.",
			Severity: hcl.DiagError,
		})
	}

	for _, build := range s.config.Builds {
		diags = diags.Extend(build.Initialize(s.config))
	}

	return diags
}

func (s *HCLSequentialScheduler) FilterBuilds(
	debug, force bool,
	onError string,
	except, only []string,
) ([]packersdk.Build, hcl.Diagnostics) {
	var allBuilds []packersdk.Build
	var diags hcl.Diagnostics

	var convertDiags hcl.Diagnostics
	s.config.Debug = debug
	s.config.Except, convertDiags = hcl2template.ConvertFilterOption(except, "except")
	diags = diags.Extend(convertDiags)
	s.config.Only, convertDiags = hcl2template.ConvertFilterOption(only, "only")
	diags = diags.Extend(convertDiags)
	s.config.Force = force
	s.config.OnError = onError

	s.config.PrepareGlobUsage()

	for _, build := range s.config.Builds {
		cbs, cbDiags := build.ToCoreBuilds(s.config)
		diags = diags.Extend(cbDiags)

		for _, cb := range cbs {
			cb.SetDebug(debug)
			cb.SetForce(force)
			cb.SetOnError(onError)

			cb.Prepared = true

			// Prepare just sets the "prepareCalled" flag on CoreBuild, since
			// we did all the prep in `ToCoreBuilds`.
			_, err := cb.Prepare()
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Preparing packer core build %s failed", cb.Name()),
					Detail:   err.Error(),
				})
			}

			allBuilds = append(allBuilds, cb)
		}
	}

	buildNames := []string{}
	for _, cb := range allBuilds {
		buildNames = append(buildNames, cb.Name())
	}

	diags = diags.Extend(
		s.config.ReportUnusedFilters(
			buildNames,
		),
	)

	return allBuilds, diags
}
