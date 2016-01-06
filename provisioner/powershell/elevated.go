package powershell

import (
	"text/template"
)

type elevatedOptions struct {
	User            string
	Password        string
	TaskName        string
	TaskDescription string
	EncodedCommand  string
}

var elevatedTemplate = template.Must(template.New("ElevatedCommand").Parse(`
$name = "{{.TaskName}}"
$log = "$env:TEMP\$name.out"
$s = New-Object -ComObject "Schedule.Service"
$s.Connect()
$t = $s.NewTask($null)
$t.XmlText = @'
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
	  <Arguments>/c powershell.exe -EncodedCommand {{.EncodedCommand}} &gt; %TEMP%\{{.TaskName}}.out 2&gt;&amp;1</Arguments>
    </Exec>
  </Actions>
</Task>
'@
$f = $s.GetFolder("\")
$f.RegisterTaskDefinition($name, $t, 6, "{{.User}}", "{{.Password}}", 1, $null) | Out-Null
$t = $f.GetTask("\$name")
$t.Run($null) | Out-Null
$timeout = 10
$sec = 0
while ((!($t.state -eq 4)) -and ($sec -lt $timeout)) {
  Start-Sleep -s 1
  $sec++
}
function SlurpOutput($l) {
  if (Test-Path $log) {
    Get-Content $log | select -skip $l | ForEach {
      $l += 1
      Write-Host "$_"
    }
  }
  return $l
}
$line = 0
do {
  Start-Sleep -m 100
  $line = SlurpOutput $line
} while (!($t.state -eq 3))
$result = $t.LastTaskResult
[System.Runtime.Interopservices.Marshal]::ReleaseComObject($s) | Out-Null
exit $result`))
