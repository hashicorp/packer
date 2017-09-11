package shell_local

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/packer/common"
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
	ExecuteCommand []string `mapstructure:"execute_command"`

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

	var errs *packer.MultiError
	if len(p.config.ExecuteCommand) == 0 {
		p.config.ExecuteCommand = []string{`/bin/sh`, `-e`, `{{.Script}}`}
	} else if len(p.config.ExecuteCommand) == 1 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("execute_command requires you to specify a slice of at least two "+
				"strings. Please see the shell-local docs for more detail."))
	}

	if p.config.Inline != nil && len(p.config.Inline) == 0 {
		p.config.Inline = nil
	}

	if p.config.InlineShebang == "" {
		p.config.InlineShebang = "/bin/sh -e"
		if runtime.GOOS == "windows" {
			p.config.InlineShebang = "sh -e"
		}
	}

	if p.config.Scripts == nil {
		p.config.Scripts = make([]string, 0)
	}

	if p.config.Vars == nil {
		p.config.Vars = make([]string, 0)
	}

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
		ui.Say(fmt.Sprintf("Making script executable"))
		if runtime.GOOS == "windows" {
			err := os.Chmod(tf.Name(), 0600) // read and write perms, which on windows is enough
			if err != nil {
				fmt.Println(err)
			}
		} else {
			err := os.Chmod(tf.Name(), 0744)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	for _, script := range scripts {

		p.config.ctx.Data = &ExecuteCommandTemplate{
			Script: script,
		}

		var interpolatedCommand []string
		for _, cmdStr := range p.config.ExecuteCommand {
			command, err := interpolate.Render(cmdStr, &p.config.ctx)
			if err != nil {
				return nil, false, fmt.Errorf("Error processing command: %s", err)
			}
			interpolatedCommand = append(interpolatedCommand, command)
		}

		ui.Say(fmt.Sprintf("Post processing with local shell script: %s", script))

		// RemoteCmd only takes a string, not a slice. To keep args separate until
		// called by exec.Command in the communicator, pass all but last element directly
		// into the communicator and only the last element of the slice to the remotecmd string.
		comm := &Communicator{
			p.config.Vars,
			interpolatedCommand[:len(interpolatedCommand)-1],
		}

		cmd := &packer.RemoteCmd{Command: interpolatedCommand[len(interpolatedCommand)-1]}

		log.Printf("starting local command: %+v", interpolatedCommand)
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
