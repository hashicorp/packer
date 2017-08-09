package command

import (
	"bufio"
	"flag"
	"fmt"
	"io"

	"github.com/hashicorp/packer/helper/flag-kv"
	"github.com/hashicorp/packer/helper/flag-slice"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template"
)

// FlagSetFlags is an enum to define what flags are present in the
// default FlagSet returned by Meta.FlagSet
type FlagSetFlags uint

const (
	FlagSetNone        FlagSetFlags = 0
	FlagSetBuildFilter FlagSetFlags = 1 << iota
	FlagSetVars
)

// Meta contains the meta-options and functionality that nearly every
// Packer command inherits.
type Meta struct {
	CoreConfig *packer.CoreConfig
	Cache      packer.Cache
	Ui         packer.Ui
	Version    string

	// These are set by command-line flags
	flagBuildExcept []string
	flagBuildOnly   []string
	flagVars        map[string]string
}

// Core returns the core for the given template given the configured
// CoreConfig and user variables on this Meta.
func (m *Meta) Core(tpl *template.Template) (*packer.Core, error) {
	// Copy the config so we don't modify it
	config := *m.CoreConfig
	config.Template = tpl
	config.Variables = m.flagVars

	// Init the core
	core, err := packer.NewCore(&config)
	if err != nil {
		return nil, fmt.Errorf("Error initializing core: %s", err)
	}

	return core, nil
}

// BuildNames returns the list of builds that are in the given core
// that we care about taking into account the only and except flags.
func (m *Meta) BuildNames(c *packer.Core) []string {
	// TODO: test

	// Filter the "only"
	if len(m.flagBuildOnly) > 0 {
		// Build a set of all the available names
		nameSet := make(map[string]struct{})
		for _, n := range c.BuildNames() {
			nameSet[n] = struct{}{}
		}

		// Build our result set which we pre-allocate some sane number
		result := make([]string, 0, len(m.flagBuildOnly))
		for _, n := range m.flagBuildOnly {
			if _, ok := nameSet[n]; ok {
				result = append(result, n)
			}
		}

		return result
	}

	// Filter the "except"
	if len(m.flagBuildExcept) > 0 {
		// Build a set of the things we don't want
		nameSet := make(map[string]struct{})
		for _, n := range m.flagBuildExcept {
			nameSet[n] = struct{}{}
		}

		// Build our result set which is the names of all builds except
		// those in the given set.
		names := c.BuildNames()
		result := make([]string, 0, len(names))
		for _, n := range names {
			if _, ok := nameSet[n]; !ok {
				result = append(result, n)
			}
		}
		return result
	}

	// We care about everything
	return c.BuildNames()
}

// FlagSet returns a FlagSet with the common flags that every
// command implements. The exact behavior of FlagSet can be configured
// using the flags as the second parameter, for example to disable
// build settings on the commands that don't handle builds.
func (m *Meta) FlagSet(n string, fs FlagSetFlags) *flag.FlagSet {
	f := flag.NewFlagSet(n, flag.ContinueOnError)

	// FlagSetBuildFilter tells us to enable the settings for selecting
	// builds we care about.
	if fs&FlagSetBuildFilter != 0 {
		f.Var((*sliceflag.StringFlag)(&m.flagBuildExcept), "except", "")
		f.Var((*sliceflag.StringFlag)(&m.flagBuildOnly), "only", "")
	}

	// FlagSetVars tells us what variables to use
	if fs&FlagSetVars != 0 {
		f.Var((*kvflag.Flag)(&m.flagVars), "var", "")
		f.Var((*kvflag.FlagJSON)(&m.flagVars), "var-file", "")
	}

	// Create an io.Writer that writes to our Ui properly for errors.
	// This is kind of a hack, but it does the job. Basically: create
	// a pipe, use a scanner to break it into lines, and output each line
	// to the UI. Do this forever.
	errR, errW := io.Pipe()
	errScanner := bufio.NewScanner(errR)
	go func() {
		for errScanner.Scan() {
			m.Ui.Error(errScanner.Text())
		}
	}()
	f.SetOutput(errW)

	return f
}

// ValidateFlags should be called after parsing flags to validate the
// given flags
func (m *Meta) ValidateFlags() error {
	// TODO
	return nil
}
