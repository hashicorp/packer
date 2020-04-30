package command

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	kvflag "github.com/hashicorp/packer/helper/flag-kv"
	sliceflag "github.com/hashicorp/packer/helper/flag-slice"
	"github.com/hashicorp/packer/helper/wrappedstreams"
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
	Ui         packer.Ui
	Version    string

	// These are set by command-line flags
	varFiles []string
	flagVars map[string]string
}

// Core returns the core for the given template given the configured
// CoreConfig and user variables on this Meta.
func (m *Meta) Core(tpl *template.Template) (*packer.Core, error) {
	// Copy the config so we don't modify it
	config := *m.CoreConfig
	config.Template = tpl

	fj := &kvflag.FlagJSON{}
	// First populate fj with contents from var files
	for _, file := range m.varFiles {
		err := fj.Set(file)
		if err != nil {
			return nil, err
		}
	}
	// Now read fj values back into flagvars and set as config.Variables. Only
	// add to flagVars if the key doesn't already exist, because flagVars comes
	// from the command line and should not be overridden by variable files.
	if m.flagVars == nil {
		m.flagVars = map[string]string{}
	}
	for k, v := range *fj {
		if _, exists := m.flagVars[k]; !exists {
			m.flagVars[k] = v
		}
	}
	config.Variables = m.flagVars

	// Init the core
	core, err := packer.NewCore(&config)
	if err != nil {
		return nil, fmt.Errorf("Error initializing core: %s", err)
	}

	return core, nil
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
		f.Var((*sliceflag.StringFlag)(&m.CoreConfig.Except), "except", "")
		f.Var((*sliceflag.StringFlag)(&m.CoreConfig.Only), "only", "")
	}

	// FlagSetVars tells us what variables to use
	if fs&FlagSetVars != 0 {
		f.Var((*kvflag.Flag)(&m.flagVars), "var", "")
		f.Var((*kvflag.StringSlice)(&m.varFiles), "var-file", "")
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

// StdinPiped returns true if the input is piped.
func (m *Meta) StdinPiped() bool {
	fi, err := wrappedstreams.Stdin().Stat()
	if err != nil {
		// If there is an error, let's just say its not piped
		return false
	}

	return fi.Mode()&os.ModeNamedPipe != 0
}
