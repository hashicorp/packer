package provisioner

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/packer"
)

type ElevatedProvisioner interface {
	Communicator() packer.Communicator
	ElevatedUser() string
	ElevatedPassword() string
}

type elevatedOptions struct {
	User              string
	Password          string
	TaskName          string
	TaskDescription   string
	LogFile           string
	XMLEscapedCommand string
	ScriptFile        string
}

var psEscape = strings.NewReplacer(
	"$", "`$",
	"\"", "`\"",
	"`", "``",
	"'", "`'",
)

var elevatedTemplate = template.Must(template.New("ElevatedCommand").Parse(`
$name = "{{.TaskName}}"
$log = [System.Environment]::ExpandEnvironmentVariables("{{.LogFile}}")
$s = New-Object -ComObject "Schedule.Service"
$s.Connect()
$t = $s.NewTask($null)
$xml = [xml]@'
<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <RegistrationInfo>
    <Description>{{.TaskDescription}}</Description>
  </RegistrationInfo>
  <Principals>
    <Principal id="Author">
      <UserId>{{.User}}</UserId>
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
      <StopOnIdleEnd>false</StopOnIdleEnd>
      <RestartOnIdle>false</RestartOnIdle>
    </IdleSettings>
    <AllowStartOnDemand>true</AllowStartOnDemand>
    <Enabled>true</Enabled>
    <Hidden>false</Hidden>
    <RunOnlyIfIdle>false</RunOnlyIfIdle>
    <WakeToRun>false</WakeToRun>
    <ExecutionTimeLimit>PT24H</ExecutionTimeLimit>
    <Priority>4</Priority>
  </Settings>
  <Actions Context="Author">
    <Exec>
      <Command>cmd</Command>
      <Arguments>/c {{.XMLEscapedCommand}}</Arguments>
    </Exec>
  </Actions>
</Task>
'@
$logon_type = 1
$password = "{{.Password}}"
if ($password.Length -eq 0) {
  $logon_type = 5
  $password = $null
  $ns = New-Object System.Xml.XmlNamespaceManager($xml.NameTable)
  $ns.AddNamespace("ns", $xml.DocumentElement.NamespaceURI)
  $node = $xml.SelectSingleNode("/ns:Task/ns:Principals/ns:Principal/ns:LogonType", $ns)
  $node.ParentNode.RemoveChild($node) | Out-Null
}
$t.XmlText = $xml.OuterXml
if (Test-Path variable:global:ProgressPreference){$ProgressPreference="SilentlyContinue"}
$f = $s.GetFolder("\")
$f.RegisterTaskDefinition($name, $t, 6, "{{.User}}", $password, $logon_type, $null) | Out-Null
$t = $f.GetTask("\$name")
$t.Run($null) | Out-Null
$timeout = 10
$sec = 0
while ((!($t.state -eq 4)) -and ($sec -lt $timeout)) {
  Start-Sleep -s 1
  $sec++
}

$line = 0
do {
  Start-Sleep -m 100
  if (Test-Path $log) {
    Get-Content $log | select -skip $line | ForEach {
      $line += 1
      Write-Output "$_"
    }
  }
} while (!($t.state -eq 3))
$result = $t.LastTaskResult
if (Test-Path $log) {
    Remove-Item $log -Force -ErrorAction SilentlyContinue | Out-Null
}

$script = [System.Environment]::ExpandEnvironmentVariables("{{.ScriptFile}}")
if (Test-Path $script) {
    Remove-Item $script -Force -ErrorAction SilentlyContinue | Out-Null
}
$f = $s.GetFolder("\")
$f.DeleteTask("\$name", "")

[System.Runtime.Interopservices.Marshal]::ReleaseComObject($s) | Out-Null
exit $result`))

func GenerateElevatedRunner(command string, p ElevatedProvisioner) (uploadedPath string, err error) {
	log.Printf("Building elevated command wrapper for: %s", command)

	var buffer bytes.Buffer

	// Output from the elevated command cannot be returned directly to the
	// Packer console. In order to be able to view output from elevated
	// commands and scripts an indirect approach is used by which the commands
	// output is first redirected to file. The output file is then 'watched'
	// by Packer while the elevated command is running and any content
	// appearing in the file is written out to the console.  Below the portion
	// of command required to redirect output from the command to file is
	// built and appended to the existing command string
	taskName := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	// Only use %ENVVAR% format for environment variables when setting the log
	// file path; Do NOT use $env:ENVVAR format as it won't be expanded
	// correctly in the elevatedTemplate
	logFile := `%SYSTEMROOT%/Temp/` + taskName + ".out"
	command += fmt.Sprintf(" > %s 2>&1", logFile)

	// elevatedTemplate wraps the command in a single quoted XML text string
	// so we need to escape characters considered 'special' in XML.
	err = xml.EscapeText(&buffer, []byte(command))
	if err != nil {
		return "", fmt.Errorf("Error escaping characters special to XML in command %s: %s", command, err)
	}
	escapedCommand := buffer.String()
	log.Printf("Command [%s] converted to [%s] for use in XML string", command, escapedCommand)
	buffer.Reset()

	// Escape chars special to PowerShell in the ElevatedUser string
	elevatedUser := p.ElevatedUser()
	escapedElevatedUser := psEscape.Replace(elevatedUser)
	if escapedElevatedUser != elevatedUser {
		log.Printf("Elevated user %s converted to %s after escaping chars special to PowerShell",
			elevatedUser, escapedElevatedUser)
	}

	// Escape chars special to PowerShell in the ElevatedPassword string
	elevatedPassword := p.ElevatedPassword()
	escapedElevatedPassword := psEscape.Replace(elevatedPassword)
	if escapedElevatedPassword != elevatedPassword {
		log.Printf("Elevated password %s converted to %s after escaping chars special to PowerShell",
			elevatedPassword, escapedElevatedPassword)
	}

	uuid := uuid.TimeOrderedUUID()
	path := fmt.Sprintf(`C:/Windows/Temp/packer-elevated-shell-%s.ps1`, uuid)

	// Generate command
	err = elevatedTemplate.Execute(&buffer, elevatedOptions{
		User:              escapedElevatedUser,
		Password:          escapedElevatedPassword,
		TaskName:          taskName,
		TaskDescription:   "Packer elevated task",
		ScriptFile:        path,
		LogFile:           logFile,
		XMLEscapedCommand: escapedCommand,
	})

	if err != nil {
		fmt.Printf("Error creating elevated template: %s", err)
		return "", err
	}
	log.Printf("Uploading elevated shell wrapper for command [%s] to [%s]", command, path)
	err = p.Communicator().Upload(path, &buffer, nil)
	if err != nil {
		return "", fmt.Errorf("Error preparing elevated powershell script: %s", err)
	}

	return fmt.Sprintf("powershell -executionpolicy bypass -file \"%s\"", path), err
}
