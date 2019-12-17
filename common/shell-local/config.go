//go:generate mapstructure-to-hcl2 -type Config

package shell_local

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/packer/common/shell"
	configHelper "github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	shell.Provisioner `mapstructure:",squash"`

	// ** DEPRECATED: USE INLINE INSTEAD **
	// ** Only Present for backwards compatibility **
	// Command is the command to execute
	Command string

	// The command used to execute the script. The '{{ .Path }}' variable
	// should be used to specify where the script goes, {{ .Vars }}
	// can be used to inject the environment_vars into the environment.
	ExecuteCommand []string `mapstructure:"execute_command"`

	// The shebang value used when running inline scripts.
	InlineShebang string `mapstructure:"inline_shebang"`

	// An array of multiple Runtime OSs to run on.
	OnlyOn []string `mapstructure:"only_on"`

	// The file extension to use for the file generated from the inline commands
	TempfileExtension string `mapstructure:"tempfile_extension"`

	// End dedupe with postprocessor
	UseLinuxPathing bool `mapstructure:"use_linux_pathing"`

	ctx interpolate.Context
}

func Decode(config *Config, raws ...interface{}) error {
	//Create passthrough for winrm password so we can fill it in once we know it
	config.ctx.Data = &EnvVarsTemplate{
		WinRMPassword: `{{.WinRMPassword}}`,
	}

	err := configHelper.Decode(config, &configHelper.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &config.ctx,
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
				"/V",
				"/C",
				"{{.Vars}}",
				"call",
				"{{.Script}}",
			}
		}
		if len(config.TempfileExtension) == 0 {
			config.TempfileExtension = ".cmd"
		}
	} else {
		if config.InlineShebang == "" {
			config.InlineShebang = "/bin/sh -e"
		}
		if len(config.ExecuteCommand) == 0 {
			config.ExecuteCommand = []string{
				"/bin/sh",
				"-c",
				"{{.Vars}} {{.Script}}",
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

	// Check for properly formatted go os types
	supportedSyslist := []string{"darwin", "freebsd", "linux", "openbsd", "solaris", "windows"}
	if len(config.OnlyOn) > 0 {
		for _, provided_os := range config.OnlyOn {
			supported_os := false
			for _, go_os := range supportedSyslist {
				if provided_os == go_os {
					supported_os = true
					break
				}
			}
			if supported_os != true {
				return fmt.Errorf("Invalid OS specified in only_on: '%s'\n"+
					"Supported OS names: %s", provided_os, strings.Join(supportedSyslist, ", "))
			}
		}
	}

	if config.UseLinuxPathing {
		for index, script := range config.Scripts {
			scriptAbsPath, err := filepath.Abs(script)
			if err != nil {
				return fmt.Errorf("Error converting %s to absolute path: %s", script, err.Error())
			}
			converted, err := ConvertToLinuxPath(scriptAbsPath)
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

	if config.EnvVarFormat == "" {
		if (runtime.GOOS == "windows") && !config.UseLinuxPathing {
			config.EnvVarFormat = "set %s=%s && "
		} else {
			config.EnvVarFormat = "%s='%s' "
		}
	}

	// drop unnecessary "." in extension; we add this later.
	if config.TempfileExtension != "" {
		if strings.HasPrefix(config.TempfileExtension, ".") {
			config.TempfileExtension = config.TempfileExtension[1:]
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
func ConvertToLinuxPath(winAbsPath string) (string, error) {
	// get absolute path of script, and morph it into the bash path
	winAbsPath = strings.Replace(winAbsPath, "\\", "/", -1)
	splitPath := strings.SplitN(winAbsPath, ":/", 2)
	if len(splitPath) == 2 {
		winBashPath := fmt.Sprintf("/mnt/%s/%s", strings.ToLower(splitPath[0]), splitPath[1])
		return winBashPath, nil
	} else {
		err := fmt.Errorf("There was an error splitting your absolute path; expected "+
			"to find a drive following the format ':/' but did not: absolute "+
			"path: %s", winAbsPath)
		return "", err
	}
}
