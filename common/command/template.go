package command

import (
	"errors"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"log"
)

// BuildOptions is a set of options related to builds that can be set
// from the command line.
type BuildOptions struct {
	UserVars map[string]string
	Except   []string
	Only     []string
}

// Validate validates the options
func (f *BuildOptions) Validate() error {
	if len(f.Except) > 0 && len(f.Only) > 0 {
		return errors.New("Only one of '-except' or '-only' may be specified.")
	}

	return nil
}

// AllUserVars returns the user variables, compiled from both the
// file paths and the vars on the command line.
func (f *BuildOptions) AllUserVars() (map[string]string, error) {
	all := make(map[string]string)

	// Copy in the command-line vars
	for k, v := range f.UserVars {
		all[k] = v
	}

	return all, nil
}

// Builds returns the builds out of the given template that pass the
// configured options.
func (f *BuildOptions) Builds(t *packer.Template, cf *packer.ComponentFinder) ([]packer.Build, error) {
	buildNames := t.BuildNames()
	builds := make([]packer.Build, 0, len(buildNames))
	for _, buildName := range buildNames {
		if len(f.Except) > 0 {
			found := false
			for _, except := range f.Except {
				if buildName == except {
					found = true
					break
				}
			}

			if found {
				log.Printf("Skipping build '%s' because specified by -except.", buildName)
				continue
			}
		}

		if len(f.Only) > 0 {
			found := false
			for _, only := range f.Only {
				if buildName == only {
					found = true
					break
				}
			}

			if !found {
				log.Printf("Skipping build '%s' because not specified by -only.", buildName)
				continue
			}
		}

		log.Printf("Creating build: %s", buildName)
		build, err := t.Build(buildName, cf)
		if err != nil {
			return nil, fmt.Errorf("Failed to create build '%s': \n\n%s", buildName, err)
		}

		builds = append(builds, build)
	}

	return builds, nil
}
