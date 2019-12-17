//go:generate mapstructure-to-hcl2 -type Config

// This package implements a provisioner for Packer that executes powershell
// scripts within the remote machine.
package powershell

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/common/shell"
	"github.com/hashicorp/packer/common/uuid"
	commonhelper "github.com/hashicorp/packer/helper/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/tmp"
	"github.com/hashicorp/packer/provisioner"
	"github.com/hashicorp/packer/template/interpolate"
)

var retryableSleep = 2 * time.Second

var psEscape = strings.NewReplacer(
	"$", "`$",
	"\"", "`\"",
	"`", "``",
	"'", "`'",
)

type Config struct {
	shell.Provisioner `mapstructure:",squash"`

	shell.ProvisionerRemoteSpecific `mapstructure:",squash"`

	// The remote path where the file containing the environment variables
	// will be uploaded to. This should be set to a writable file that is in a
	// pre-existing directory.
	RemoteEnvVarPath string `mapstructure:"remote_env_var_path"`

	// The command used to execute the elevated script. The '{{ .Path }}'
	// variable should be used to specify where the script goes, {{ .Vars }}
	// can be used to inject the environment_vars into the environment.
	ElevatedExecuteCommand string `mapstructure:"elevated_execute_command"`

	// The timeout for retrying to start the process. Until this timeout is
	// reached, if the provisioner can't start a process, it retries.  This
	// can be set high to allow for reboots.
	StartRetryTimeout time.Duration `mapstructure:"start_retry_timeout"`

	// This is used in the template generation to format environment variables
	// inside the `ElevatedExecuteCommand` template.
	ElevatedEnvVarFormat string `mapstructure:"elevated_env_var_format"`

	// Instructs the communicator to run the remote script as a Windows
	// scheduled task, effectively elevating the remote user by impersonating
	// a logged-in user
	ElevatedUser     string `mapstructure:"elevated_user"`
	ElevatedPassword string `mapstructure:"elevated_password"`

	ExecutionPolicy ExecutionPolicy `mapstructure:"execution_policy"`

	ctx interpolate.Context
}

type Provisioner struct {
	config       Config
	communicator packer.Communicator
}

type ExecuteCommandTemplate struct {
	Vars          string
	Path          string
	WinRMPassword string
}

type EnvVarsTemplate struct {
	WinRMPassword string
}

func (p *Provisioner) defaultExecuteCommand() string {
	baseCmd := `& { if (Test-Path variable:global:ProgressPreference)` +
		`{set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};` +
		`. {{.Vars}}; &'{{.Path}}'; exit $LastExitCode }`
	if p.config.ExecutionPolicy == ExecutionPolicyNone {
		return baseCmd
	} else {
		return fmt.Sprintf(`powershell -executionpolicy %s "%s"`, p.config.ExecutionPolicy, baseCmd)
	}
}
func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	// Create passthrough for winrm password so we can fill it in once we know
	// it
	p.config.ctx.Data = &EnvVarsTemplate{
		WinRMPassword: `{{.WinRMPassword}}`,
	}

	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
				"elevated_execute_command",
			},
		},
		DecodeHooks: append(config.DefaultDecodeHookFuncs, StringToExecutionPolicyHook),
	}, raws...)

	if err != nil {
		return err
	}

	if p.config.EnvVarFormat == "" {
		p.config.EnvVarFormat = `$env:%s="%s"; `
	}

	if p.config.ElevatedEnvVarFormat == "" {
		p.config.ElevatedEnvVarFormat = `$env:%s="%s"; `
	}

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = p.defaultExecuteCommand()
	}

	if p.config.ElevatedExecuteCommand == "" {
		p.config.ElevatedExecuteCommand = p.defaultExecuteCommand()
	}

	if p.config.Inline != nil && len(p.config.Inline) == 0 {
		p.config.Inline = nil
	}

	if p.config.StartRetryTimeout == 0 {
		p.config.StartRetryTimeout = 5 * time.Minute
	}

	if p.config.RemotePath == "" {
		uuid := uuid.TimeOrderedUUID()
		p.config.RemotePath = fmt.Sprintf(`c:/Windows/Temp/script-%s.ps1`, uuid)
	}

	if p.config.RemoteEnvVarPath == "" {
		uuid := uuid.TimeOrderedUUID()
		p.config.RemoteEnvVarPath = fmt.Sprintf(`c:/Windows/Temp/packer-ps-env-vars-%s.ps1`, uuid)
	}

	if p.config.Scripts == nil {
		p.config.Scripts = make([]string, 0)
	}

	if p.config.Vars == nil {
		p.config.Vars = make([]string, 0)
	}

	var errs error
	if p.config.Script != "" && len(p.config.Scripts) > 0 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Only one of script or scripts can be specified."))
	}

	if p.config.ElevatedUser == "" && p.config.ElevatedPassword != "" {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Must supply an 'elevated_user' if 'elevated_password' provided"))
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

	if errs != nil {
		return errs
	}

	return nil
}

// Takes the inline scripts, concatenates them into a temporary file and
// returns a string containing the location of said file.
func extractScript(p *Provisioner) (string, error) {
	temp, err := tmp.File("powershell-provisioner")
	if err != nil {
		return "", err
	}
	defer temp.Close()
	writer := bufio.NewWriter(temp)
	for _, command := range p.config.Inline {
		log.Printf("Found command: %s", command)
		if _, err := writer.WriteString(command + "\n"); err != nil {
			return "", fmt.Errorf("Error preparing powershell script: %s", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("Error preparing powershell script: %s", err)
	}

	return temp.Name(), nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	ui.Say(fmt.Sprintf("Provisioning with Powershell..."))
	p.communicator = comm

	scripts := make([]string, len(p.config.Scripts))
	copy(scripts, p.config.Scripts)

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
		ui.Say(fmt.Sprintf("Provisioning with powershell script: %s", path))

		log.Printf("Opening %s for reading", path)
		fi, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("Error stating powershell script: %s", err)
		}
		if strings.HasSuffix(p.config.RemotePath, `\`) {
			// path is a directory
			p.config.RemotePath += filepath.Base((fi).Name())
		}
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Error opening powershell script: %s", err)
		}
		defer f.Close()

		command, err := p.createCommandText()
		if err != nil {
			return fmt.Errorf("Error processing command: %s", err)
		}

		// Upload the file and run the command. Do this in the context of a
		// single retryable function so that we don't end up with the case
		// that the upload succeeded, a restart is initiated, and then the
		// command is executed but the file doesn't exist any longer.
		var cmd *packer.RemoteCmd
		err = retry.Config{StartTimeout: p.config.StartRetryTimeout}.Run(ctx, func(ctx context.Context) error {
			if _, err := f.Seek(0, 0); err != nil {
				return err
			}
			if err := comm.Upload(p.config.RemotePath, f, &fi); err != nil {
				return fmt.Errorf("Error uploading script: %s", err)
			}

			cmd = &packer.RemoteCmd{Command: command}
			return cmd.RunWithUi(ctx, comm, ui)
		})
		if err != nil {
			return err
		}

		// Close the original file since we copied it
		f.Close()

		log.Printf("%s returned with exit code %d", p.config.RemotePath, cmd.ExitStatus())

		if err := p.config.ValidExitCode(cmd.ExitStatus()); err != nil {
			return err
		}
	}

	return nil
}

// Environment variables required within the remote environment are uploaded
// within a PS script and then enabled by 'dot sourcing' the script
// immediately prior to execution of the main command
func (p *Provisioner) prepareEnvVars(elevated bool) (err error) {
	// Collate all required env vars into a plain string with required
	// formatting applied
	flattenedEnvVars := p.createFlattenedEnvVars(elevated)
	// Create a powershell script on the target build fs containing the
	// flattened env vars
	err = p.uploadEnvVars(flattenedEnvVars)
	if err != nil {
		return err
	}
	return
}

func (p *Provisioner) createFlattenedEnvVars(elevated bool) (flattened string) {
	flattened = ""
	envVars := make(map[string]string)

	// Always available Packer provided env vars
	envVars["PACKER_BUILD_NAME"] = p.config.PackerBuildName
	envVars["PACKER_BUILDER_TYPE"] = p.config.PackerBuilderType

	// expose ip address variables
	httpAddr := common.GetHTTPAddr()
	if httpAddr != "" {
		envVars["PACKER_HTTP_ADDR"] = httpAddr
	}
	httpIP := common.GetHTTPIP()
	if httpIP != "" {
		envVars["PACKER_HTTP_IP"] = httpIP
	}
	httpPort := common.GetHTTPPort()
	if httpPort != "" {
		envVars["PACKER_HTTP_PORT"] = httpPort
	}

	// interpolate environment variables
	p.config.ctx.Data = &EnvVarsTemplate{
		WinRMPassword: getWinRMPassword(p.config.PackerBuildName),
	}
	// Split vars into key/value components
	for _, envVar := range p.config.Vars {
		envVar, err := interpolate.Render(envVar, &p.config.ctx)
		if err != nil {
			return
		}
		keyValue := strings.SplitN(envVar, "=", 2)
		// Escape chars special to PS in each env var value
		escapedEnvVarValue := psEscape.Replace(keyValue[1])
		if escapedEnvVarValue != keyValue[1] {
			log.Printf("Env var %s converted to %s after escaping chars special to PS", keyValue[1],
				escapedEnvVarValue)
		}
		envVars[keyValue[0]] = escapedEnvVarValue
	}

	// Create a list of env var keys in sorted order
	var keys []string
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	format := p.config.EnvVarFormat
	if elevated {
		format = p.config.ElevatedEnvVarFormat
	}

	// Re-assemble vars using OS specific format pattern and flatten
	for _, key := range keys {
		flattened += fmt.Sprintf(format, key, envVars[key])
	}
	return
}

func (p *Provisioner) uploadEnvVars(flattenedEnvVars string) (err error) {
	ctx := context.TODO()
	// Upload all env vars to a powershell script on the target build file
	// system. Do this in the context of a single retryable function so that
	// we gracefully handle any errors created by transient conditions such as
	// a system restart
	envVarReader := strings.NewReader(flattenedEnvVars)
	log.Printf("Uploading env vars to %s", p.config.RemoteEnvVarPath)
	err = retry.Config{StartTimeout: p.config.StartRetryTimeout}.Run(ctx, func(context.Context) error {
		if err := p.communicator.Upload(p.config.RemoteEnvVarPath, envVarReader, nil); err != nil {
			return fmt.Errorf("Error uploading ps script containing env vars: %s", err)
		}
		return err
	})
	if err != nil {
		return err
	}
	return
}

func (p *Provisioner) createCommandText() (command string, err error) {
	// Return the interpolated command
	if p.config.ElevatedUser == "" {
		return p.createCommandTextNonPrivileged()
	} else {
		return p.createCommandTextPrivileged()
	}
}

func (p *Provisioner) createCommandTextNonPrivileged() (command string, err error) {
	// Prepare everything needed to enable the required env vars within the
	// remote environment
	err = p.prepareEnvVars(false)
	if err != nil {
		return "", err
	}

	p.config.ctx.Data = &ExecuteCommandTemplate{
		Path:          p.config.RemotePath,
		Vars:          p.config.RemoteEnvVarPath,
		WinRMPassword: getWinRMPassword(p.config.PackerBuildName),
	}
	command, err = interpolate.Render(p.config.ExecuteCommand, &p.config.ctx)

	if err != nil {
		return "", fmt.Errorf("Error processing command: %s", err)
	}

	// Return the interpolated command
	return command, nil
}

func getWinRMPassword(buildName string) string {
	winRMPass, _ := commonhelper.RetrieveSharedState("winrm_password", buildName)
	packer.LogSecretFilter.Set(winRMPass)
	return winRMPass
}

func (p *Provisioner) createCommandTextPrivileged() (command string, err error) {
	// Prepare everything needed to enable the required env vars within the
	// remote environment
	err = p.prepareEnvVars(true)
	if err != nil {
		return "", err
	}

	p.config.ctx.Data = &ExecuteCommandTemplate{
		Path:          p.config.RemotePath,
		Vars:          p.config.RemoteEnvVarPath,
		WinRMPassword: getWinRMPassword(p.config.PackerBuildName),
	}
	command, err = interpolate.Render(p.config.ElevatedExecuteCommand, &p.config.ctx)
	if err != nil {
		return "", fmt.Errorf("Error processing command: %s", err)
	}

	command, err = provisioner.GenerateElevatedRunner(command, p)
	if err != nil {
		return "", fmt.Errorf("Error generating elevated runner: %s", err)
	}

	return command, err
}

func (p *Provisioner) Communicator() packer.Communicator {
	return p.communicator
}

func (p *Provisioner) ElevatedUser() string {
	return p.config.ElevatedUser
}

func (p *Provisioner) ElevatedPassword() string {
	// Replace ElevatedPassword for winrm users who used this feature
	p.config.ctx.Data = &EnvVarsTemplate{
		WinRMPassword: getWinRMPassword(p.config.PackerBuildName),
	}

	elevatedPassword, _ := interpolate.Render(p.config.ElevatedPassword, &p.config.ctx)

	return elevatedPassword
}
