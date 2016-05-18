package shell_local

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

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

	// The command used to execute the script. The '{{ .Path }}' variable
	// should be used to specify where the script goes, {{ .Vars }}
	// can be used to inject the environment_vars into the environment.
	ExecuteCommand string `mapstructure:"execute_command"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

type ExecuteCommandTemplate struct {
	Vars     string
	Script   string
	Artifact string
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = "chmod +x {{.Script}}; {{.Vars}} {{.Script}} {{.Artifact}}"
	}

	if p.config.Inline != nil && len(p.config.Inline) == 0 {
		p.config.Inline = nil
	}

	if p.config.InlineShebang == "" {
		p.config.InlineShebang = "/bin/sh -e"
	}

	if p.config.Scripts == nil {
		p.config.Scripts = make([]string, 0)
	}

	if p.config.Vars == nil {
		p.config.Vars = make([]string, 0)
	}

	var errs *packer.MultiError
	if p.config.Script != "" && len(p.config.Scripts) > 0 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Only one of script or scripts can be specified."))
	}

	if p.config.Script != "" {
		p.config.Scripts = []string{p.config.Script}
	}

	if len(p.config.Scripts) == 0 && p.config.Inline == nil {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Either a script file or inline script must be specified."))
	} else if len(p.config.Scripts) > 0 && p.config.Inline != nil {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Only a script file or an inline script can be specified, not both."))
	}

	for _, path := range p.config.Scripts {
		if _, err := os.Stat(path); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Bad script '%s': %s", path, err))
		}
	}

	// Do a check for bad environment variables, such as '=foo', 'foobar'
	for idx, kv := range p.config.Vars {
		vs := strings.SplitN(kv, "=", 2)
		if len(vs) != 2 || vs[0] == "" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Environment variable not in format 'key=value': %s", kv))
		} else {
			// Replace single quotes so they parse
			vs[1] = strings.Replace(vs[1], "'", `'"'"'`, -1)

			// Single quote env var values
			p.config.Vars[idx] = fmt.Sprintf("%s='%s'", vs[0], vs[1])
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {

	scripts := make([]string, len(p.config.Scripts))
	copy(scripts, p.config.Scripts)

	// If we have an inline script, then turn that into a temporary
	// shell script and use that.
	if p.config.Inline != nil {
		tf, err := ioutil.TempFile("", "packer-shell")
		if err != nil {
			return nil, false, fmt.Errorf("Error preparing shell script: %s", err)
		}
		defer os.Remove(tf.Name())

		// Set the path to the temporary file
		scripts = append(scripts, tf.Name())

		// Write our contents to it
		writer := bufio.NewWriter(tf)
		writer.WriteString(fmt.Sprintf("#!%s\n", p.config.InlineShebang))
		for _, command := range p.config.Inline {
			if _, err := writer.WriteString(command + "\n"); err != nil {
				return nil, false, fmt.Errorf("Error preparing shell script: %s", err)
			}
		}

		if err := writer.Flush(); err != nil {
			return nil, false, fmt.Errorf("Error preparing shell script: %s", err)
		}

		tf.Close()
	}

	// Build our variables up by adding in the build name and builder type
	envVars := make([]string, len(p.config.Vars)+2)
	envVars[0] = fmt.Sprintf("PACKER_BUILD_NAME='%s'", p.config.PackerBuildName)
	envVars[1] = fmt.Sprintf("PACKER_BUILDER_TYPE='%s'", p.config.PackerBuilderType)
	copy(envVars[2:], p.config.Vars)

	for _, file := range artifact.Files() {
		for _, script := range scripts {
			// Flatten the environment variables
			flattendVars := strings.Join(envVars, " ")

			p.config.ctx.Data = &ExecuteCommandTemplate{
				Vars:     flattendVars,
				Script:   script,
				Artifact: file,
			}

			command, err := interpolate.Render(p.config.ExecuteCommand, &p.config.ctx)
			if err != nil {
				return nil, false, fmt.Errorf("Error processing command: %s", err)
			}

			ui.Say(fmt.Sprintf("Post processing with local shell script: %s", command))

			comm := &Communicator{}

			cmd := &packer.RemoteCmd{Command: command}

			ui.Say(fmt.Sprintf(
				"Executing local script: %s",
				script))
			if err := cmd.StartWithUi(comm, ui); err != nil {
				return nil, false, fmt.Errorf(
					"Error executing script: %s\n\n"+
						"Please see output above for more information.",
					script)
			}
			if cmd.ExitStatus != 0 {
				return nil, false, fmt.Errorf(
					"Erroneous exit code %d while executing script: %s\n\n"+
						"Please see output above for more information.",
					cmd.ExitStatus,
					script)
			}
		}
	}

	return artifact, true, nil
}
