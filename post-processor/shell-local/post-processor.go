package shell_local

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/packer/common"
	sl "github.com/hashicorp/packer/common/shell-local"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
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
	Vars   string
	Script string
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
		p.config.ExecuteCommand = `chmod +x "{{.Script}}"; {{.Vars}} "{{.Script}}"`
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
	for _, kv := range p.config.Vars {
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

	// Create environment variables to set before executing the command
	flattenedEnvVars := p.createFlattenedEnvVars()

	for _, script := range scripts {

		p.config.ctx.Data = &ExecuteCommandTemplate{
			Vars:   flattenedEnvVars,
			Script: script,
		}

		command, err := interpolate.Render(p.config.ExecuteCommand, &p.config.ctx)
		if err != nil {
			return nil, false, fmt.Errorf("Error processing command: %s", err)
		}

		ui.Say(fmt.Sprintf("Post processing with local shell script: %s", script))

		comm := &sl.Communicator{
			Ctx:            p.config.ctx,
			ExecuteCommand: []string{p.config.ExecuteCommand},
		}

		cmd := &packer.RemoteCmd{Command: command}

		log.Printf("starting local command: %s", command)
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

	return artifact, true, nil
}

func (p *PostProcessor) createFlattenedEnvVars() (flattened string) {
	flattened = ""
	envVars := make(map[string]string)

	// Always available Packer provided env vars
	envVars["PACKER_BUILD_NAME"] = fmt.Sprintf("%s", p.config.PackerBuildName)
	envVars["PACKER_BUILDER_TYPE"] = fmt.Sprintf("%s", p.config.PackerBuilderType)

	// Split vars into key/value components
	for _, envVar := range p.config.Vars {
		keyValue := strings.SplitN(envVar, "=", 2)
		// Store pair, replacing any single quotes in value so they parse
		// correctly with required environment variable format
		envVars[keyValue[0]] = strings.Replace(keyValue[1], "'", `'"'"'`, -1)
	}

	// Create a list of env var keys in sorted order
	var keys []string
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Re-assemble vars surrounding value with single quotes and flatten
	for _, key := range keys {
		flattened += fmt.Sprintf("%s='%s' ", key, envVars[key])
	}
	return
}
