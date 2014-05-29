package winrm

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	isotime "github.com/mitchellh/packer/common/time"
	"github.com/mitchellh/packer/packer"
	"github.com/sneal/go-winrm"
)

type comm struct {
	client   *winrm.Client
	address  string
	user     string
	password string
	timeout  time.Duration
}

type elevatedShellOptions struct {
	Command  string
	User     string
	Password string
}

// Creates a new packer.Communicator implementation over WinRM.
// Called when Packer tries to connect to WinRM
func New(address string, user string, password string, timeout time.Duration) (*comm, error) {

	// Create the WinRM client we use internally
	params := winrm.DefaultParameters()
	params.Timeout = isotime.ISO8601DurationString(timeout)
	client := winrm.NewClientWithParameters(address, user, password, params)

	// Attempt to connect to the WinRM service
	shell, err := client.CreateShell()
	if err != nil {
		return nil, err
	}

	err = shell.Close()
	if err != nil {
		return nil, err
	}

	return &comm{
		address:  address,
		user:     user,
		password: password,
		timeout:  timeout,
		client:   client,
	}, nil
}

func (c *comm) Start(cmd *packer.RemoteCmd) (err error) {
	return c.StartElevated(cmd)
}

func (c *comm) StartElevated(cmd *packer.RemoteCmd) (err error) {
	// Wrap the command in scheduled task
	tpl, err := packer.NewConfigTemplate()
	if err != nil {
		return err
	}

	elevatedScript, err := tpl.Process(ElevatedShellTemplate, &elevatedShellOptions{
		Command:  cmd.Command,
		User:     c.user,
		Password: c.password,
	})
	if err != nil {
		return err
	}

	// Upload the script which creates and manages the scheduled task
	err = c.Upload("$env:TEMP/packer-elevated-shell.ps1", strings.NewReader(elevatedScript))
	if err != nil {
		return err
	}

	// Run the script that was uploaded
	command := fmt.Sprintf("powershell -executionpolicy bypass -file \"%s\"", "%TEMP%/packer-elevated-shell.ps1")
	return c.runCommand(command, cmd)
}

func (c *comm) StartUnelevated(cmd *packer.RemoteCmd) (err error) {
	return c.runCommand(cmd.Command, cmd)
}

func (c *comm) runCommand(commandText string, cmd *packer.RemoteCmd) (err error) {
	log.Printf("starting remote command: %s", cmd.Command)

	// Create a new shell process on the guest
	shell, err := c.client.CreateShell()
	if err != nil {
		log.Printf("error creating shell, retrying once more: %s", err)
		shell, err = c.client.CreateShell()
		if err != nil {
			log.Printf("error creating shell, giving up: %s", err)
			return err
		}
	}

	// Execute the command
	var winrmCmd *winrm.Command
	winrmCmd, err = shell.Execute(commandText, cmd.Stdout, cmd.Stderr)
	if err != nil {
		log.Printf("error executing command: %s", err)
		return err
	}

	// Start a goroutine to wait for the shell to end and set the
	// exit boolean and status.
	go func() {

		defer func() {
			err = winrmCmd.Close()
			if err != nil {
				log.Printf("winrm command failed to close: %s", err)
			}
			err = shell.Close()
			if err != nil {
				log.Printf("shell failed to close: %s", err)
			}
		}()

		// Block until done
		winrmCmd.Wait()

		// Report exit status and trigger done
		exitStatus := winrmCmd.ExitCode()
		log.Printf("remote command exited with '%d'", exitStatus)
		cmd.SetExited(exitStatus)
	}()

	return
}

func (c *comm) Upload(dst string, input io.Reader) error {
	fm := &fileManager{
		comm: c,
	}
	return fm.Upload(dst, input)
}

func (c *comm) UploadDir(dst string, src string, excl []string) error {
	fm := &fileManager{
		comm: c,
	}
	return fm.UploadDir(dst, src)
}

func (c *comm) Download(string, io.Writer) error {
	panic("Download not implemented yet")
}

const ElevatedShellTemplate = `
$command = "{{.Command}}" + '; exit $LASTEXITCODE'
$user = '{{.User}}'
$password = '{{.Password}}'

$task_name = "packer-elevated-shell"
$out_file = "$env:TEMP\packer-elevated-shell.log"

if (Test-Path $out_file) {
  del $out_file
}

$task_xml = @'
<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <Principals>
    <Principal id="Author">
      <UserId>{user}</UserId>
      <LogonType>Password</LogonType>
      <RunLevel>HighestAvailable</RunLevel>
    </Principal>
  </Principals>
  <Settings>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
    <AllowHardTerminate>true</AllowHardTerminate>
    <StartWhenAvailable>false</StartWhenAvailable>
    <RunOnlyIfNetworkAvailable>false</RunOnlyIfNetworkAvailable>
    <IdleSettings>
      <StopOnIdleEnd>true</StopOnIdleEnd>
      <RestartOnIdle>false</RestartOnIdle>
    </IdleSettings>
    <AllowStartOnDemand>true</AllowStartOnDemand>
    <Enabled>true</Enabled>
    <Hidden>false</Hidden>
    <RunOnlyIfIdle>false</RunOnlyIfIdle>
    <WakeToRun>false</WakeToRun>
    <ExecutionTimeLimit>PT2H</ExecutionTimeLimit>
    <Priority>4</Priority>
  </Settings>
  <Actions Context="Author">
    <Exec>
      <Command>cmd</Command>
      <Arguments>{arguments}</Arguments>
    </Exec>
  </Actions>
</Task>
'@

$bytes = [System.Text.Encoding]::Unicode.GetBytes($command)
$encoded_command = [Convert]::ToBase64String($bytes)
$arguments = "/c powershell.exe -EncodedCommand $encoded_command &gt; $out_file 2&gt;&amp;1"

$task_xml = $task_xml.Replace("{arguments}", $arguments)
$task_xml = $task_xml.Replace("{user}", $user)

$schedule = New-Object -ComObject "Schedule.Service"
$schedule.Connect()
$task = $schedule.NewTask($null)
$task.XmlText = $task_xml
$folder = $schedule.GetFolder("\")
$folder.RegisterTaskDefinition($task_name, $task, 6, $user, $password, 1, $null) | Out-Null

$registered_task = $folder.GetTask("\$task_name")
$registered_task.Run($null) | Out-Null

$timeout = 10
$sec = 0
while ( (!($registered_task.state -eq 4)) -and ($sec -lt $timeout) ) {
  Start-Sleep -s 1
  $sec++
}

function SlurpOutput($out_file, $cur_line) {
  if (Test-Path $out_file) {
    get-content $out_file | select -skip $cur_line | ForEach {
      $cur_line += 1
      Write-Host "$_" 
    }
  }
  return $cur_line
}

$cur_line = 0
do {
  Start-Sleep -m 100
  $cur_line = SlurpOutput $out_file $cur_line
} while (!($registered_task.state -eq 3))

$exit_code = $registered_task.LastTaskResult
[System.Runtime.Interopservices.Marshal]::ReleaseComObject($schedule) | Out-Null

exit $exit_code
`
