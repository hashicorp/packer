package shell_local

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/packer/common"
	configHelper "github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// ** DEPRECATED: USE INLINE INSTEAD **
	// ** Only Present for backwards compatibiltiy **
	// Command is the command to execute
	Command string

	// An inline script to execute. Multiple strings are all executed
	// in the context of a single shell.
	Inline []string

	// The shebang value used when running inline scripts.
	InlineShebang string `mapstructure:"inline_shebang"`

	// The local path of the shell script to upload and execute.
	Script string

	// An array of multiple scripts to run.
	Scripts []string

	// An array of environment variables that will be injected before
	// your command(s) are executed.
	Vars []string `mapstructure:"environment_vars"`
	// End dedupe with postprocessor

	// The command used to execute the script. The '{{ .Path }}' variable
	// should be used to specify where the script goes, {{ .Vars }}
	// can be used to inject the environment_vars into the environment.
	ExecuteCommand []string `mapstructure:"execute_command"`

	Ctx interpolate.Context
}

func Decode(config *Config, raws ...interface{}) error {
	err := configHelper.Decode(&config, &configHelper.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &config.Ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	return Validate(config)
}

func Validate(config *Config) error {
	var errs *packer.MultiError

	if runtime.GOOS == "windows" {
		if config.InlineShebang == "" {
			config.InlineShebang = ""
		}
		if len(config.ExecuteCommand) == 0 {
			config.ExecuteCommand = []string{`{{.Vars}} "{{.Script}}"`}
		}
	} else {
		if config.InlineShebang == "" {
			// TODO: verify that provisioner defaulted to this as well
			config.InlineShebang = "/bin/sh -e"
		}
		if len(config.ExecuteCommand) == 0 {
			config.ExecuteCommand = []string{`chmod +x "{{.Script}}"; {{.Vars}} "{{.Script}}"`}
		}
	}

	// Clean up input
	if config.Inline != nil && len(config.Inline) == 0 {
		config.Inline = nil
	}

	if config.Scripts == nil {
		config.Scripts = make([]string, 0)
	}

	if config.Vars == nil {
		config.Vars = make([]string, 0)
	}

	// Verify that the user has given us a command to run
	if config.Command != "" && len(config.Inline) == 0 &&
		len(config.Scripts) == 0 && config.Script == "" {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Command, Inline, Script and Scripts options cannot all be empty."))
	}

	if config.Command != "" {
		// Backwards Compatibility: Before v1.2.2, the shell-local
		// provisioner only allowed a single Command, and to run
		// multiple commands you needed to run several provisioners in a
		// row, one for each command. In deduplicating the post-processor and
		// provisioner code, we've changed this to allow an array of scripts or
		// inline commands just like in the post-processor. This conditional
		// grandfathers in the "Command" option, allowing the original usage to
		// continue to work.
		config.Inline = append(config.Inline, config.Command)
	}

	if config.Script != "" && len(config.Scripts) > 0 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Only one of script or scripts can be specified."))
	}

	if config.Script != "" {
		config.Scripts = []string{config.Script}
	}

	if len(config.Scripts) > 0 && config.Inline != nil {
		errs = packer.MultiErrorAppend(errs,
			errors.New("You may specify either a script file(s) or an inline script(s), but not both."))
	}

	for _, path := range config.Scripts {
		if _, err := os.Stat(path); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Bad script '%s': %s", path, err))
		}
	}

	// Do a check for bad environment variables, such as '=foo', 'foobar'
	for _, kv := range config.Vars {
		vs := strings.SplitN(kv, "=", 2)
		if len(vs) != 2 || vs[0] == "" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Environment variable not in format 'key=value': %s", kv))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}
