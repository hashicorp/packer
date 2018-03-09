package shell_local

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

	UseLinuxPathing bool `mapstructure:"use_linux_pathing"`

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
		return fmt.Errorf("Error decoding config: %s, config is %#v, and raws is %#v", err, config, raws)
	}

	return nil
}

func Validate(config *Config) error {
	var errs *packer.MultiError

	if runtime.GOOS == "windows" {
		if len(config.ExecuteCommand) == 0 {
			config.ExecuteCommand = []string{
				"cmd",
				"/C",
				"{{.Vars}}",
				"{{.Script}}",
			}
		}
	} else {
		if config.InlineShebang == "" {
			config.InlineShebang = "/bin/sh -e"
		}
		if len(config.ExecuteCommand) == 0 {
			config.ExecuteCommand = []string{
				"/bin/sh",
				"-c",
				"{{.Vars}}",
				"{{.Script}}",
			}
		}
	}

	// Clean up input
	if config.Inline != nil && len(config.Inline) == 0 {
		config.Inline = make([]string, 0)
	}

	if config.Scripts == nil {
		config.Scripts = make([]string, 0)
	}

	if config.Vars == nil {
		config.Vars = make([]string, 0)
	}

	// Verify that the user has given us a command to run
	if config.Command == "" && len(config.Inline) == 0 &&
		len(config.Scripts) == 0 && config.Script == "" {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Command, Inline, Script and Scripts options cannot all be empty."))
	}

	// Check that user hasn't given us too many commands to run
	tooManyOptionsErr := errors.New("You may only specify one of the " +
		"following options: Command, Inline, Script or Scripts. Please" +
		" consolidate these options in your config.")

	if config.Command != "" {
		if len(config.Inline) != 0 || len(config.Scripts) != 0 || config.Script != "" {
			errs = packer.MultiErrorAppend(errs, tooManyOptionsErr)
		} else {
			config.Inline = []string{config.Command}
		}
	}

	if config.Script != "" {
		if len(config.Scripts) > 0 || len(config.Inline) > 0 {
			errs = packer.MultiErrorAppend(errs, tooManyOptionsErr)
		} else {
			config.Scripts = []string{config.Script}
		}
	}

	if len(config.Scripts) > 0 && config.Inline != nil {
		errs = packer.MultiErrorAppend(errs, tooManyOptionsErr)
	}

	// Check that all scripts we need to run exist locally
	for _, path := range config.Scripts {
		if _, err := os.Stat(path); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Bad script '%s': %s", path, err))
		}
	}
	if config.UseLinuxPathing {
		for index, script := range config.Scripts {
			converted, err := ConvertToLinuxPath(script)
			if err != nil {
				return err
			}
			config.Scripts[index] = converted
		}
		// Interoperability issues with WSL makes creating and running tempfiles
		// via golang's os package basically impossible.
		if len(config.Inline) > 0 {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Packer is unable to use the Command and Inline "+
					"features with the Windows Linux Subsystem. Please use "+
					"the Script or Scripts options instead"))
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

// C:/path/to/your/file becomes /mnt/c/path/to/your/file
func ConvertToLinuxPath(winPath string) (string, error) {
	// get absolute path of script, and morph it into the bash path
	winAbsPath, err := filepath.Abs(winPath)
	if err != nil {
		return "", fmt.Errorf("Error converting %s to absolute path: %s", winPath, err.Error())
	}
	winAbsPath = strings.Replace(winAbsPath, "\\", "/", -1)
	splitPath := strings.SplitN(winAbsPath, ":/", 2)
	winBashPath := fmt.Sprintf("/mnt/%s/%s", strings.ToLower(splitPath[0]), splitPath[1])
	return winBashPath, nil
}
