//go:generate mapstructure-to-hcl2 -type Config

// This package implements a provisioner for Packer that executes
// shell scripts within the remote machine.
package shell

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/common/shell"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/tmp"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	shell.Provisioner `mapstructure:",squash"`

	shell.ProvisionerRemoteSpecific `mapstructure:",squash"`

	// The shebang value used when running inline scripts.
	InlineShebang string `mapstructure:"inline_shebang"`

	// A duration of how long to pause after the provisioner
	PauseAfter time.Duration `mapstructure:"pause_after"`

	// Write the Vars to a file and source them from there rather than declaring
	// inline
	UseEnvVarFile bool `mapstructure:"use_env_var_file"`

	// The remote folder where the local shell script will be uploaded to.
	// This should be set to a pre-existing directory, it defaults to /tmp
	RemoteFolder string `mapstructure:"remote_folder"`

	// The remote file name of the local shell script.
	// This defaults to script_nnn.sh
	RemoteFile string `mapstructure:"remote_file"`

	// The timeout for retrying to start the process. Until this timeout
	// is reached, if the provisioner can't start a process, it retries.
	// This can be set high to allow for reboots.
	StartRetryTimeout time.Duration `mapstructure:"start_retry_timeout"`

	// Whether to clean scripts up
	SkipClean bool `mapstructure:"skip_clean"`

	ExpectDisconnect bool `mapstructure:"expect_disconnect"`

	// name of the tmp environment variable file, if UseEnvVarFile is true
	envVarFile string

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

type ExecuteCommandTemplate struct {
	Vars       string
	EnvVarFile string
	Path       string
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
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

	if p.config.EnvVarFormat == "" {
		p.config.EnvVarFormat = "%s='%s' "

		if p.config.UseEnvVarFile == true {
			p.config.EnvVarFormat = "export %s='%s'\n"
		}
	}

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = "chmod +x {{.Path}}; {{.Vars}} {{.Path}}"
		if p.config.UseEnvVarFile == true {
			p.config.ExecuteCommand = "chmod +x {{.Path}}; . {{.EnvVarFile}} && {{.Path}}"
		}
	}

	if p.config.Inline != nil && len(p.config.Inline) == 0 {
		p.config.Inline = nil
	}

	if p.config.InlineShebang == "" {
		p.config.InlineShebang = "/bin/sh -e"
	}

	if p.config.StartRetryTimeout == 0 {
		p.config.StartRetryTimeout = 5 * time.Minute
	}

	if p.config.RemoteFolder == "" {
		p.config.RemoteFolder = "/tmp"
	}

	if p.config.RemoteFile == "" {
		p.config.RemoteFile = fmt.Sprintf("script_%d.sh", rand.Intn(9999))
	}

	if p.config.RemotePath == "" {
		p.config.RemotePath = fmt.Sprintf(
			"%s/%s", p.config.RemoteFolder, p.config.RemoteFile)
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

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	scripts := make([]string, len(p.config.Scripts))
	copy(scripts, p.config.Scripts)

	// If we have an inline script, then turn that into a temporary
	// shell script and use that.
	if p.config.Inline != nil {
		tf, err := tmp.File("packer-shell")
		if err != nil {
			return fmt.Errorf("Error preparing shell script: %s", err)
		}
		defer os.Remove(tf.Name())

		// Set the path to the temporary file
		scripts = append(scripts, tf.Name())

		// Write our contents to it
		writer := bufio.NewWriter(tf)
		writer.WriteString(fmt.Sprintf("#!%s\n", p.config.InlineShebang))
		for _, command := range p.config.Inline {
			if _, err := writer.WriteString(command + "\n"); err != nil {
				return fmt.Errorf("Error preparing shell script: %s", err)
			}
		}

		if err := writer.Flush(); err != nil {
			return fmt.Errorf("Error preparing shell script: %s", err)
		}

		tf.Close()
	}

	if p.config.UseEnvVarFile == true {
		tf, err := tmp.File("packer-shell-vars")
		if err != nil {
			return fmt.Errorf("Error preparing shell script: %s", err)
		}
		defer os.Remove(tf.Name())

		// Write our contents to it
		writer := bufio.NewWriter(tf)
		if _, err := writer.WriteString(p.createEnvVarFileContent()); err != nil {
			return fmt.Errorf("Error preparing shell script: %s", err)
		}

		if err := writer.Flush(); err != nil {
			return fmt.Errorf("Error preparing shell script: %s", err)
		}

		p.config.envVarFile = tf.Name()
		defer os.Remove(p.config.envVarFile)

		// upload the var file
		var cmd *packer.RemoteCmd
		err = retry.Config{StartTimeout: p.config.StartRetryTimeout}.Run(ctx, func(ctx context.Context) error {
			if _, err := tf.Seek(0, 0); err != nil {
				return err
			}

			var r io.Reader = tf
			if !p.config.Binary {
				r = &UnixReader{Reader: r}
			}
			remoteVFName := fmt.Sprintf("%s/%s", p.config.RemoteFolder,
				fmt.Sprintf("varfile_%d.sh", rand.Intn(9999)))
			if err := comm.Upload(remoteVFName, r, nil); err != nil {
				return fmt.Errorf("Error uploading envVarFile: %s", err)
			}
			tf.Close()

			cmd = &packer.RemoteCmd{
				Command: fmt.Sprintf("chmod 0600 %s", remoteVFName),
			}
			if err := comm.Start(ctx, cmd); err != nil {
				return fmt.Errorf(
					"Error chmodding script file to 0600 in remote "+
						"machine: %s", err)
			}
			cmd.Wait()
			p.config.envVarFile = remoteVFName
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Create environment variables to set before executing the command
	flattenedEnvVars := p.createFlattenedEnvVars()

	for _, path := range scripts {
		ui.Say(fmt.Sprintf("Provisioning with shell script: %s", path))

		log.Printf("Opening %s for reading", path)
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Error opening shell script: %s", err)
		}
		defer f.Close()

		// Compile the command
		p.config.ctx.Data = &ExecuteCommandTemplate{
			Vars:       flattenedEnvVars,
			EnvVarFile: p.config.envVarFile,
			Path:       p.config.RemotePath,
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
		var cmd *packer.RemoteCmd
		err = retry.Config{StartTimeout: p.config.StartRetryTimeout}.Run(ctx, func(ctx context.Context) error {
			if _, err := f.Seek(0, 0); err != nil {
				return err
			}

			var r io.Reader = f
			if !p.config.Binary {
				r = &UnixReader{Reader: r}
			}

			if err := comm.Upload(p.config.RemotePath, r, nil); err != nil {
				return fmt.Errorf("Error uploading script: %s", err)
			}

			cmd = &packer.RemoteCmd{
				Command: fmt.Sprintf("chmod 0755 %s", p.config.RemotePath),
			}
			if err := comm.Start(ctx, cmd); err != nil {
				return fmt.Errorf(
					"Error chmodding script file to 0755 in remote "+
						"machine: %s", err)
			}
			cmd.Wait()

			cmd = &packer.RemoteCmd{Command: command}
			return cmd.RunWithUi(ctx, comm, ui)
		})

		if err != nil {
			return err
		}

		// If the exit code indicates a remote disconnect, fail unless
		// we were expecting it.
		if cmd.ExitStatus() == packer.CmdDisconnect {
			if !p.config.ExpectDisconnect {
				return fmt.Errorf("Script disconnected unexpectedly. " +
					"If you expected your script to disconnect, i.e. from a " +
					"restart, you can try adding `\"expect_disconnect\": true` " +
					"or `\"valid_exit_codes\": [0, 2300218]` to the shell " +
					"provisioner parameters.")
			}
		} else if err := p.config.ValidExitCode(cmd.ExitStatus()); err != nil {
			return err
		}

		if !p.config.SkipClean {

			// Delete the temporary file we created. We retry this a few times
			// since if the above rebooted we have to wait until the reboot
			// completes.
			err = p.cleanupRemoteFile(p.config.RemotePath, comm)
			if err != nil {
				return err
			}
			err = p.cleanupRemoteFile(p.config.envVarFile, comm)
			if err != nil {
				return err
			}
		}
	}

	if p.config.PauseAfter != 0 {
		ui.Say(fmt.Sprintf("Pausing %s after this provisioner...", p.config.PauseAfter))
		select {
		case <-time.After(p.config.PauseAfter):
			return nil
		}
	}

	return nil
}

func (p *Provisioner) cleanupRemoteFile(path string, comm packer.Communicator) error {
	ctx := context.TODO()
	err := retry.Config{StartTimeout: p.config.StartRetryTimeout}.Run(ctx, func(ctx context.Context) error {
		cmd := &packer.RemoteCmd{
			Command: fmt.Sprintf("rm -f %s", path),
		}
		if err := comm.Start(ctx, cmd); err != nil {
			return fmt.Errorf(
				"Error removing temporary script at %s: %s",
				path, err)
		}
		cmd.Wait()
		// treat disconnects as retryable by returning an error
		if cmd.ExitStatus() == packer.CmdDisconnect {
			return fmt.Errorf("Disconnect while removing temporary script.")
		}
		if cmd.ExitStatus() != 0 {
			return fmt.Errorf(
				"Error removing temporary script at %s!",
				path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) escapeEnvVars() ([]string, map[string]string) {
	envVars := make(map[string]string)

	// Always available Packer provided env vars
	envVars["PACKER_BUILD_NAME"] = fmt.Sprintf("%s", p.config.PackerBuildName)
	envVars["PACKER_BUILDER_TYPE"] = fmt.Sprintf("%s", p.config.PackerBuilderType)

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

	return keys, envVars
}

func (p *Provisioner) createEnvVarFileContent() string {
	keys, envVars := p.escapeEnvVars()

	var flattened string
	for _, key := range keys {
		flattened += fmt.Sprintf(p.config.EnvVarFormat, key, envVars[key])
	}

	return flattened
}

func (p *Provisioner) createFlattenedEnvVars() string {
	keys, envVars := p.escapeEnvVars()

	// Re-assemble vars into specified format and flatten
	var flattened string
	for _, key := range keys {
		flattened += fmt.Sprintf(p.config.EnvVarFormat, key, envVars[key])
	}

	return flattened
}
