//go:generate packer-sdc mapstructure-to-hcl2 -type Config

// This package implements a provisioner for Packer that executes
// shell scripts within the remote machine.
package shell

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	"github.com/hashicorp/packer-plugin-sdk/shell"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
)

// FIXME query remote host or use %SYSTEMROOT%, %TEMP% and more creative filename
const DefaultRemotePath = "c:/Windows/Temp/script.bat"

type Config struct {
	shell.Provisioner `mapstructure:",squash"`

	shell.ProvisionerRemoteSpecific `mapstructure:",squash"`

	// The timeout for retrying to start the process. Until this timeout
	// is reached, if the provisioner can't start a process, it retries.
	// This can be set high to allow for reboots.
	StartRetryTimeout time.Duration `mapstructure:"start_retry_timeout"`

	ctx interpolate.Context
}

type Provisioner struct {
	config        Config
	generatedData map[string]interface{}
}

type ExecuteCommandTemplate struct {
	Vars string
	Path string
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "windows-shell",
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

	if p.config.EnvVarFormat == "" {
		p.config.EnvVarFormat = `set "%s=%s" && `
	}

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = `{{.Vars}}"{{.Path}}"`
	}

	if p.config.Inline != nil && len(p.config.Inline) == 0 {
		p.config.Inline = nil
	}

	if p.config.StartRetryTimeout == 0 {
		p.config.StartRetryTimeout = 5 * time.Minute
	}

	if p.config.RemotePath == "" {
		p.config.RemotePath = DefaultRemotePath
	}

	if p.config.Scripts == nil {
		p.config.Scripts = make([]string, 0)
	}

	if p.config.Vars == nil {
		p.config.Vars = make([]string, 0)
	}

	var errs error
	if p.config.Script != "" && len(p.config.Scripts) > 0 {
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("Only one of script or scripts can be specified."))
	}

	if p.config.Script != "" {
		p.config.Scripts = []string{p.config.Script}
	}

	if len(p.config.Scripts) == 0 && p.config.Inline == nil {
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("Either a script file or inline script must be specified."))
	} else if len(p.config.Scripts) > 0 && p.config.Inline != nil {
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("Only a script file or an inline script can be specified, not both."))
	}

	for _, path := range p.config.Scripts {
		if _, err := os.Stat(path); err != nil {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("Bad script '%s': %s", path, err))
		}
	}

	// Do a check for bad environment variables, such as '=foo', 'foobar'
	for _, kv := range p.config.Vars {
		vs := strings.SplitN(kv, "=", 2)
		if len(vs) != 2 || vs[0] == "" {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("Environment variable not in format 'key=value': %s", kv))
		}
	}

	return errs
}

// This function takes the inline scripts, concatenates them
// into a temporary file and returns a string containing the location
// of said file.
func extractScript(p *Provisioner) (string, error) {
	temp, err := tmp.File("windows-shell-provisioner")
	if err != nil {
		log.Printf("Unable to create temporary file for inline scripts: %s", err)
		return "", err
	}
	writer := bufio.NewWriter(temp)
	for _, command := range p.config.Inline {
		log.Printf("Found command: %s", command)
		if _, err := writer.WriteString(command + "\n"); err != nil {
			return "", fmt.Errorf("Error preparing shell script: %s", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("Error preparing shell script: %s", err)
	}

	temp.Close()

	return temp.Name(), nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{}) error {
	ui.Say("Provisioning with windows-shell...")
	scripts := make([]string, len(p.config.Scripts))
	copy(scripts, p.config.Scripts)
	p.generatedData = generatedData

	if p.config.Inline != nil {
		temp, err := extractScript(p)
		if err != nil {
			ui.Error(fmt.Sprintf("Unable to extract inline scripts into a file: %s", err))
		}
		scripts = append(scripts, temp)
		// Remove temp script containing the inline commands when done
		defer os.Remove(temp)
	}

	for _, path := range scripts {
		ui.Say(fmt.Sprintf("Provisioning with shell script: %s", path))

		log.Printf("Opening %s for reading", path)
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Error opening shell script: %s", err)
		}
		defer f.Close()

		// Create environment variables to set before executing the command
		flattenedVars := p.createFlattenedEnvVars()

		// Compile the command
		p.config.ctx.Data = &ExecuteCommandTemplate{
			Vars: flattenedVars,
			Path: p.config.RemotePath,
		}
		command, err := interpolate.Render(p.config.ExecuteCommand, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Error processing command: %s", err)
		}

		// Upload the file and run the command. Do this in the context of
		// a single retryable function so that we don't end up with
		// the case that the upload succeeded, a restart is initiated,
		// and then the command is executed but the file doesn't exist
		// any longer.
		var cmd *packersdk.RemoteCmd
		err = retry.Config{StartTimeout: p.config.StartRetryTimeout}.Run(ctx, func(ctx context.Context) error {
			if _, err := f.Seek(0, 0); err != nil {
				return err
			}

			if err := comm.Upload(p.config.RemotePath, f, nil); err != nil {
				return fmt.Errorf("Error uploading script: %s", err)
			}

			cmd = &packersdk.RemoteCmd{Command: command}
			return cmd.RunWithUi(ctx, comm, ui)
		})
		if err != nil {
			return err
		}

		// Close the original file since we copied it
		f.Close()

		if err := p.config.ValidExitCode(cmd.ExitStatus()); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provisioner) createFlattenedEnvVars() (flattened string) {
	flattened = ""
	envVars := make(map[string]string)

	// Always available Packer provided env vars
	envVars["PACKER_BUILD_NAME"] = p.config.PackerBuildName
	envVars["PACKER_BUILDER_TYPE"] = p.config.PackerBuilderType

	// expose ip address variables
	httpAddr := p.generatedData["PackerHTTPAddr"]
	if httpAddr != nil && httpAddr != commonsteps.HttpAddrNotImplemented {
		envVars["PACKER_HTTP_ADDR"] = httpAddr.(string)
	}
	httpIP := p.generatedData["PackerHTTPIP"]
	if httpIP != nil && httpIP != commonsteps.HttpIPNotImplemented {
		envVars["PACKER_HTTP_IP"] = httpIP.(string)
	}
	httpPort := p.generatedData["PackerHTTPPort"]
	if httpPort != nil && httpPort != commonsteps.HttpPortNotImplemented {
		envVars["PACKER_HTTP_PORT"] = httpPort.(string)
	}

	// Split vars into key/value components
	for _, envVar := range p.config.Vars {
		keyValue := strings.SplitN(envVar, "=", 2)
		envVars[keyValue[0]] = keyValue[1]
	}

	for k, v := range p.config.Env {
		envVars[k] = v
	}

	// Create a list of env var keys in sorted order
	var keys []string
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// Re-assemble vars using OS specific format pattern and flatten
	for _, key := range keys {
		flattened += fmt.Sprintf(p.config.EnvVarFormat, key, envVars[key])
	}
	return
}
