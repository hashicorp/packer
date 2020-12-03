package command

import (
	"bufio"
	"flag"
	"io"
	"os"

	kvflag "github.com/hashicorp/packer/command/flag-kv"
	"github.com/hashicorp/packer/helper/wrappedstreams"
	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template"
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
	Ui         packersdk.Ui
	Version    string
}

// Core returns the core for the given template given the configured
// CoreConfig and user variables on this Meta.
func (m *Meta) Core(tpl *template.Template, cla *MetaArgs) (*packer.Core, error) {
	// Copy the config so we don't modify it
	config := *m.CoreConfig
	config.Template = tpl

	fj := &kvflag.FlagJSON{}
	// First populate fj with contents from var files
	for _, file := range cla.VarFiles {
		err := fj.Set(file)
		if err != nil {
			return nil, err
		}
	}
	// Now read fj values back into flagvars and set as config.Variables. Only
	// add to flagVars if the key doesn't already exist, because flagVars comes
	// from the command line and should not be overridden by variable files.
	if cla.Vars == nil {
		cla.Vars = map[string]string{}
	}
	for k, v := range *fj {
		if _, exists := cla.Vars[k]; !exists {
			cla.Vars[k] = v
		}
	}
	config.Variables = cla.Vars

	core := packer.NewCore(&config)
	return core, nil
}

// FlagSet returns a FlagSet with the common flags that every
// command implements. The exact behavior of FlagSet can be configured
// using the flags as the second parameter, for example to disable
// build settings on the commands that don't handle builds.
func (m *Meta) FlagSet(n string, _ FlagSetFlags) *flag.FlagSet {
	f := flag.NewFlagSet(n, flag.ContinueOnError)

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
