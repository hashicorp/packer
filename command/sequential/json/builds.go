package json

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func (s *JSONSequentialScheduler) PrepareBuilds() hcl.Diagnostics {
	var diags hcl.Diagnostics

	// Go through and interpolate all the build names. We should be able
	// to do this at this point with the variables.
	s.config.Builds = make(map[string]*template.Builder)
	for _, b := range s.config.Template.Builders {
		v, err := interpolate.Render(b.Name, s.config.Context())
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Build interpolation failure",
				Detail: fmt.Sprintf("Error interpolating builder '%s': %s",
					b.Name, err),
			})
		}

		s.config.Builds[v] = b
	}

	return diags
}

// This is used for json templates to launch the build plugins.
// They will be prepared via b.Prepare() later.
func (s *JSONSequentialScheduler) FilterBuilds(
	debug, force bool,
	onError string,
	except, only []string,
) ([]packersdk.Build, hcl.Diagnostics) {
	buildNames := s.config.BuildNames(only, except)
	builds := []packersdk.Build{}
	diags := hcl.Diagnostics{}
	for _, n := range buildNames {
		b, err := s.config.Build(n)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Failed to initialize build %q", n),
				Detail:   err.Error(),
			})
			continue
		}

		// Now that build plugin has been launched, call Prepare()
		log.Printf("Preparing build: %s", b.Name())
		b.SetDebug(debug)
		b.SetForce(force)
		b.SetOnError(onError)

		warnings, err := b.Prepare()
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Failed to prepare build: %q", n),
				Detail:   err.Error(),
			})
			continue
		}

		// Only append builds to list if the Prepare() is successful.
		builds = append(builds, b)

		if len(warnings) > 0 {
			for _, warning := range warnings {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  fmt.Sprintf("Warning when preparing build: %q", n),
					Detail:   warning,
				})
			}
		}
	}
	return builds, diags
}
