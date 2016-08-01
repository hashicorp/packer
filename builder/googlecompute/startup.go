package googlecompute

import (
	"encoding/base64"
	"fmt"
)

const StartupScriptStartLog string = "Packer startup script starting."
const StartupScriptDoneLog string = "Packer startup script done."
const StartupScriptKey string = "startup-script"
const StartupWrappedScriptKey string = "packer-wrapped-startup-script"

const StartupWindowsSysprepKey string = "sysprep-specialize-script-ps1"

// We have to encode StartupScriptDoneLog because we use it as a sentinel value to indicate
// that the user-provided startup script is done. If we pass StartupScriptDoneLog as-is, it
// will be printed early in the instance console log (before the startup script even runs;
// we print out instance creation metadata which contains this wrapper script).
var StartupScriptDoneLogBase64 string = base64.StdEncoding.EncodeToString([]byte(StartupScriptDoneLog))

var StartupScript string = fmt.Sprintf(`#!/bin/bash
echo %s
RETVAL=0

GetMetadata () {
	echo "$(curl -f -H "Metadata-Flavor: Google" http://metadata/computeMetadata/v1/instance/attributes/$1 2> /dev/null)"
}

STARTUPSCRIPT=$(GetMetadata %s)
STARTUPSCRIPTPATH=/packer-wrapped-startup-script
if [ -f "/var/log/startupscript.log" ]; then
  STARTUPSCRIPTLOGPATH=/var/log/startupscript.log
else
  STARTUPSCRIPTLOGPATH=/var/log/daemon.log
fi
STARTUPSCRIPTLOGDEST=$(GetMetadata startup-script-log-dest)

if [[ ! -z $STARTUPSCRIPT ]]; then
  echo "Executing user-provided startup script..."
  echo "${STARTUPSCRIPT}" > ${STARTUPSCRIPTPATH}
  chmod +x ${STARTUPSCRIPTPATH}
  ${STARTUPSCRIPTPATH}
  RETVAL=$?

  if [[ ! -z $STARTUPSCRIPTLOGDEST ]]; then
    echo "Uploading user-provided startup script log to ${STARTUPSCRIPTLOGDEST}..."
    gsutil -h "Content-Type:text/plain" cp ${STARTUPSCRIPTLOGPATH} ${STARTUPSCRIPTLOGDEST}
  fi

  rm ${STARTUPSCRIPTPATH}
fi

echo $(echo %s | base64 --decode)
exit $RETVAL
`, StartupScriptStartLog, StartupWrappedScriptKey, StartupScriptDoneLogBase64)

// Pulled parts of the Ansible WinRM script and the future GoogleComputeEngine script

var StartupWinRMScript = `
# Import Modules
try {
  Import-Module 'C:\Program Files\Google\Compute Engine\sysprep\gce_base.psm1' -ErrorAction Stop
}
catch [System.Management.Automation.ActionPreferenceStopException] {
  Write-Host $_.Exception.GetBaseException().Message
  Write-Host ("Unable to import GCE module from C:\Program Files\Google\Compute Engine\sysprep. " +
      'Check error message, or ensure module is present.')
  exit 2
}

Function New-LegacySelfSignedCert
{
    Param (
        [string]$SubjectName,
        [int]$ValidDays = 365,
        [string]$CertStoreLocation = "Cert:\LocalMachine\my"
    )

    $name = New-Object -COM "X509Enrollment.CX500DistinguishedName.1"
    $name.Encode("CN=$SubjectName", 0)

    $key = New-Object -COM "X509Enrollment.CX509PrivateKey.1"
    $key.ProviderName = "Microsoft RSA SChannel Cryptographic Provider"
    $key.KeySpec = 1
    $key.Length = 1024
    $key.SecurityDescriptor = "D:PAI(A;;0xd01f01ff;;;SY)(A;;0xd01f01ff;;;BA)(A;;0x80120089;;;NS)"
    $key.MachineContext = 1
    $key.Create()

    $serverauthoid = New-Object -COM "X509Enrollment.CObjectId.1"
    $serverauthoid.InitializeFromValue("1.3.6.1.5.5.7.3.1")
    $ekuoids = New-Object -COM "X509Enrollment.CObjectIds.1"
    $ekuoids.Add($serverauthoid)
    $ekuext = New-Object -COM "X509Enrollment.CX509ExtensionEnhancedKeyUsage.1"
    $ekuext.InitializeEncode($ekuoids)

    $cert = New-Object -COM "X509Enrollment.CX509CertificateRequestCertificate.1"
    $cert.InitializeFromPrivateKey(2, $key, "")
    $cert.Subject = $name
    $cert.Issuer = $cert.Subject
    $cert.NotBefore = (Get-Date).AddDays(-1)
    $cert.NotAfter = $cert.NotBefore.AddDays($ValidDays)
    $cert.X509Extensions.Add($ekuext)
    $cert.Encode()

    $enrollment = New-Object -COM "X509Enrollment.CX509Enrollment.1"
    $enrollment.InitializeFromRequest($cert)
    $certdata = $enrollment.CreateRequest(0)
    $enrollment.InstallResponse(2, $certdata, 0, "")

    # Return the thumbprint of the last installed certificate;
    # This is needed for the new HTTPS WinRM listerner we're
    # going to create further down.
    Get-ChildItem $CertStoreLocation | Sort-Object NotBefore -Descending | Select -First 1
}

function Get-InstanceName {
  
  Write-Log 'Getting hostname from metadata server.'

  if ((Get-CimInstance Win32_BIOS).Manufacturer -cne 'Google') {
    Write-Log 'Not running in a Google Compute Engine VM.' -error
    return
  }

  $count = 1
  do {
    $hostname_parts = (_FetchFromMetaData -property 'hostname') -split '\.'
    if ($hostname_parts.Length -le 1) {
      Write-Log "Waiting for metadata server, attempt $count."
      Start-Sleep -Seconds 1
    }
    if ($count++ -ge 60) {
      Write-Log 'There is likely a problem with the network.' -error
      return
    }
  }
  while ($hostname_parts.Length -le 1)

  $hostname_parts[0]
  
}

function Configure-WinRM {
  <#
    .SYNOPSIS
      Setup WinRM on the instance.
    .DESCRIPTION
      Create a self signed cert to use with a HTTPS WinRM endpoint and restart the WinRM service.
  #>

  Write-Log 'Configuring WinRM...'

  # Running before a reboot hostname won't be correct so get it from metadata server
  
  $name = Get-InstanceName
  # We're using makecert here because New-SelfSignedCertificate isn't full featured in anything
  # less than Win10/Server 2016, makecert is installed during imaging on non 2016 machines.
  try {
    $cert = New-SelfSignedCertificate -DnsName "$($name)" -CertStoreLocation 'Cert:\LocalMachine\My'
  }
  catch {
    $cert = New-LegacySelfSignedCert -SubjectName "$($name)"
  }
  # Configure winrm HTTPS transport using the created cert.
  $config = '@{Hostname="'+ $($name) + '";CertificateThumbprint="' + $cert.Thumbprint + '";port="5986"}'
  _RunExternalCMD winrm create winrm/config/listener?Address=*+Transport=HTTPS $config
  # Open the firewall.

  # Check for basic authentication.
    $basicAuthSetting = Get-ChildItem WSMan:\localhost\Service\Auth | Where {$_.Name -eq "Basic"}
    If (($basicAuthSetting.Value) -eq $false)
    {
        Write-Verbose "Enabling basic auth support."
        Set-Item -Path "WSMan:\localhost\Service\Auth\Basic" -Value $true
    }
    Else
    {
        Write-Verbose "Basic auth is already enabled."
    }
  $rule = 'Windows Remote Management (HTTPS-In)'
  _RunExternalCMD netsh advfirewall firewall add rule profile=any name=$rule dir=in localport=5986 protocol=TCP action=allow

  Restart-Service WinRM
  Write-Log 'Setup of WinRM complete.'
}
$httpsOptions = New-PSSessionOption -SkipCACheck -SkipCNCheck -SkipRevocationCheck

$httpsResult = New-PSSession -UseSSL -ComputerName "localhost" -SessionOption $httpsOptions -ErrorVariable httpsError -ErrorAction SilentlyContinue

if(!$httpsResult) {
    Configure-WinRM
}
`
