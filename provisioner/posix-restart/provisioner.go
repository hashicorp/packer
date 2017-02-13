package restart

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

var DefaultRestartCommands = []string{
	"if [ -x /sbin/systemctl -o -x /usr/sbin/systemctl ];then",
	"nohup sh -c 'systemctl stop sshd && reboot'",
	"exit $?",
	"fi",
	"if [ -x /sbin/service -o -x /usr/sbin/service ];then",
	"nohup sh -c 'service sshd stop  && shutdown -r now'",
	"exit $?",
	"fi",
	"if [ -x /etc/init.d/sshd ];then",
	"nohup sh -c '/etc/init.d/sshd stop  && shutdown -r now'",
	"exit $?",
	"fi",
	"echo 'ERROR: I do not know how to restart this machine'",
	"exit 1",
}
var DefaultRestartCheckCommand = "echo \"$(hostname) restarted.\""
var retryableSleep = 5 * time.Second

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The command used to execute the script. The '{{ .Path }}' variable
	// should be used to specify where the script goes, {{ .Vars }}
	// can be used to inject the environment_vars into the environment.
	ExecuteCommand string `mapstructure:"execute_command"`

	// The remote path of the script
	// This should be set to a writable file that is in a pre-existing directory.
	// Internally used
	RemotePath string `mapstructure:"remote_path"`

	// Commands to restart the guest machine. Multiple strings are all executed
	// in the context of a single shell.
	RestartCommands []string `mapstructure:"restart_commands"`

	// The command used to check if the guest machine has restarted
	// The output of this command will be displayed to the user
	RestartCheckCommand string `mapstructure:"restart_check_command"`

	// The timeout for waiting for the machine to restart
	RestartTimeout time.Duration `mapstructure:"restart_timeout"`

	// The shebang value used when running the generated script.
	Shebang string `mapstructure:"shebang"`

	// Whether to clean scripts up
	SkipClean bool `mapstructure:"skip_clean"`

	ctx interpolate.Context
}

type Provisioner struct {
	config     Config
	comm       packer.Communicator
	ui         packer.Ui
	cancel     chan struct{}
	cancelLock sync.Mutex
}

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

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = "chmod +x {{.Path}}; {{.Vars}} sudo -S -E sh {{.Path}}"
	}

	if p.config.RestartCommands != nil && len(p.config.RestartCommands) == 0 {
		p.config.RestartCommands = DefaultRestartCommands
	}

	if p.config.RestartCheckCommand == "" {
		p.config.RestartCheckCommand = DefaultRestartCheckCommand
	}

	if p.config.RestartTimeout == 0 {
		p.config.RestartTimeout = 5 * time.Minute
	}

	if p.config.Shebang == "" {
		p.config.Shebang = "/bin/sh -e"
	}

	return nil
}

type ExecuteCommandTemplate struct {
	Vars string
	Path string
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	p.cancelLock.Lock()
	p.cancel = make(chan struct{})
	p.cancelLock.Unlock()

	ui.Say("Restarting Machine")
	p.comm = comm
	p.ui = ui

	tf, err := ioutil.TempFile("", "packer-shell")
	if err != nil {
		return fmt.Errorf("Error preparing shell script: %s", err)
	}
	log.Printf("Preparing restart shell script: %s", tf.Name())
	defer os.Remove(tf.Name())

	// Write our contents to it
	writer := bufio.NewWriter(tf)
	writer.WriteString(fmt.Sprintf("#!%s\n", p.config.Shebang))
	for _, command := range p.config.RestartCommands {
		if _, err := writer.WriteString(command + "\n"); err != nil {
			return fmt.Errorf("Error preparing shell script: %s", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Error preparing shell script: %s", err)
	}

	tf.Close()

	log.Printf("Opening %s for reading", tf.Name())
	f, err := os.Open(tf.Name())
	if err != nil {
		return fmt.Errorf("Error opening shell script: %s", err)
	}
	defer f.Close()

	// Create environment variables to set before executing the command
	flattenedEnvVars := p.createFlattenedEnvVars()

	// Compile the command
	p.config.ctx.Data = &ExecuteCommandTemplate{
		Vars: flattenedEnvVars,
		Path: tf.Name(),
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
	err = p.retryable(func() error {
		if _, err := f.Seek(0, 0); err != nil {
			return err
		}

		if err := comm.Upload(tf.Name(), f, nil); err != nil {
			return fmt.Errorf("Error uploading script: %s", err)
		}

		cmd = &packer.RemoteCmd{
			Command: fmt.Sprintf("chmod 0755 %s", tf.Name()),
		}
		if err := comm.Start(cmd); err != nil {
			return fmt.Errorf(
				"Error chmodding script file to 0755 in remote "+
					"machine: %s", err)
		}
		p.config.RemotePath = tf.Name()
		cmd.Wait()

		cmd = &packer.RemoteCmd{Command: command}
		return cmd.StartWithUi(comm, ui)
	})

	if err != nil {
		return err
	}

	// The exit code can indicates a remote disconnect it is normal for restart provisionner
	if cmd.ExitStatus != packer.CmdDisconnect && cmd.ExitStatus != 0 {
		return fmt.Errorf("Script exited with non-zero exit status: %d", cmd.ExitStatus)
	}

	return waitForRestart(p, comm)
}

var waitForRestart = func(p *Provisioner, comm packer.Communicator) error {
	ui := p.ui
	ui.Say("Waiting for machine to restart...")
	waitDone := make(chan bool, 1)
	timeout := time.After(p.config.RestartTimeout)
	var err error

	go func() {
		log.Printf("Waiting for machine to become available...")
		err = waitForCommunicator(p)
		waitDone <- true
	}()

	log.Printf("Waiting for machine to reboot with timeout: %s", p.config.RestartTimeout)

WaitLoop:
	for {
		// Wait for either SSH to become available, a timeout to occur,
		// or an interrupt to come through.
		select {
		case <-waitDone:
			if err != nil {
				ui.Error(fmt.Sprintf("Error waiting for SSH: %s", err))
				return err
			}

			ui.Say("Machine successfully restarted, moving on")
			//			close(p.cancel)
			break WaitLoop
		case <-timeout:
			err := fmt.Errorf("Timeout waiting for SSH.")
			ui.Error(err.Error())
			close(p.cancel)
			return err
		case <-p.cancel:
			close(waitDone)
			return fmt.Errorf("Interrupt detected, quitting waiting for machine to restart")
		}
	}

	waitClean := make(chan bool, 1)
	go func() {
		log.Printf("Waiting for temporary script to be removed...")
		err = CleanTemporary(p)
		waitClean <- true
	}()

	log.Printf("Waiting for temporary script removal ...")
CleanLoop:
	for {
		// Wait for either Files to be removed
		// or an interrupt to come through.
		select {
		case <-waitClean:
			if err != nil {
				ui.Error(fmt.Sprintf("Error waiting for Cleaning: %s", err))
				return err
			}
			log.Printf("Temporary script successfully removed, moving on")
			close(p.cancel)
			break CleanLoop
		case <-p.cancel:
			close(waitClean)
			return fmt.Errorf("Interrupt detected, quitting waiting for temporary script to be removed")
		}
	}

	return nil

}

var waitForCommunicator = func(p *Provisioner) error {
	cmd := &packer.RemoteCmd{Command: p.config.RestartCheckCommand}

	for {
		select {
		case <-p.cancel:
			log.Println("Communicator wait cancelled, exiting loop")
			return fmt.Errorf("Communicator wait cancelled")
		case <-time.After(retryableSleep):
		}

		log.Printf("Attempting to communicator to machine with: '%s'", cmd.Command)

		err := cmd.StartWithUi(p.comm, p.ui)
		if err != nil {
			log.Printf("Communication connection err: %s", err)
			continue
		}
		log.Printf("Connected to machine")
		break
	}
	return nil
}

var CleanTemporary = func(p *Provisioner) error {
	if !p.config.SkipClean {
		log.Printf("Attempting to remove temporary script: '%s'", p.config.RemotePath)
		// Delete the temporary file we created.
		cmd := &packer.RemoteCmd{
			Command: fmt.Sprintf("rm -f %s", p.config.RemotePath),
		}
		if err := p.comm.Start(cmd); err != nil {
			return fmt.Errorf(
				"Error removing temporary script at %s: %s",
				p.config.RemotePath, err)
		}
		cmd.Wait()
		// treat disconnects as retryable by returning an error
		if cmd.ExitStatus == packer.CmdDisconnect {
			return fmt.Errorf("Disconnect while removing temporary script.")
		}

		if cmd.ExitStatus != 0 {
			return fmt.Errorf("Error removing temporary script at %s!", p.config.RemotePath)
		}
		log.Printf("Temporary script removed: '%s'", p.config.RemotePath)
	}
	return nil
}

func (p *Provisioner) Cancel() {
	log.Printf("Received interrupt Cancel()")

	p.cancelLock.Lock()
	defer p.cancelLock.Unlock()
	if p.cancel != nil {
		close(p.cancel)
	}
}

// retryable will retry the given function over and over until a
// non-error is returned.
func (p *Provisioner) retryable(f func() error) error {
	startTimeout := time.After(p.config.RestartTimeout)
	for {
		var err error
		if err = f(); err == nil {
			return nil
		}

		// Create an error and log it
		err = fmt.Errorf("Retryable error: %s", err)
		log.Printf(err.Error())

		// Check if we timed out, otherwise we retry. It is safe to
		// retry since the only error case above is if the command
		// failed to START.
		select {
		case <-startTimeout:
			return err
		default:
			time.Sleep(retryableSleep)
		}
	}
}
func (p *Provisioner) createFlattenedEnvVars() (flattened string) {
	flattened = ""
	envVars := make(map[string]string)

	// Always available Packer provided env vars
	envVars["PACKER_BUILD_NAME"] = fmt.Sprintf("%s", p.config.PackerBuildName)
	envVars["PACKER_BUILDER_TYPE"] = fmt.Sprintf("%s", p.config.PackerBuilderType)
	httpAddr := common.GetHTTPAddr()
	if httpAddr != "" {
		envVars["PACKER_HTTP_ADDR"] = fmt.Sprintf("%s", httpAddr)
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
