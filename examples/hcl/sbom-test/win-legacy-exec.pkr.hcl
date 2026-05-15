packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.8"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "windows" {
  ami_name      = "test1-windows-legacy-exec"
  instance_type = "m4.2xlarge"
  region        = "us-west-2"

  source_ami_filter {
    filters = {
      name                = "Windows_Server-2022-English-Full-Base-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["801119661308"]
  }

  communicator   = "winrm"
  winrm_username = "packer"
  winrm_password = "P@cker1234"
  winrm_insecure = true
  winrm_use_ssl  = true
  winrm_timeout  = "30m"

  user_data = <<EOF
<powershell>

Set-ExecutionPolicy Unrestricted -Scope LocalMachine -Force -ErrorAction Ignore
$ErrorActionPreference = "stop"

Remove-Item -Path WSMan:\Localhost\listener\listener* -Recurse
$Cert = New-SelfSignedCertificate -CertstoreLocation Cert:\LocalMachine\My -DnsName "packer"
New-Item -Path WSMan:\LocalHost\Listener -Transport HTTPS -Address * -CertificateThumbPrint $Cert.Thumbprint -Force

$username = "packer"
$password = ConvertTo-SecureString -String 'P@cker1234' -AsPlainText -Force
New-LocalUser $username -Password $password -FullName "Packer User" -Description "Temporary user for Packer"
Add-LocalGroupMember -Group "Administrators" -Member $username

cmd.exe /c winrm quickconfig -q
cmd.exe /c winrm set "winrm/config" '@{MaxTimeoutms="1800000"}'
cmd.exe /c winrm set "winrm/config" '@{MaxEnvelopeSizekb="8192"}'
cmd.exe /c winrm set "winrm/config/winrs" '@{MaxMemoryPerShellMB="1024"}'
cmd.exe /c winrm set "winrm/config/service" '@{AllowUnencrypted="true"}'
cmd.exe /c winrm set "winrm/config/client" '@{AllowUnencrypted="true"}'
cmd.exe /c winrm set "winrm/config/service/auth" '@{Basic="true"}'
cmd.exe /c winrm set "winrm/config/client/auth" '@{Basic="true"}'
cmd.exe /c winrm set "winrm/config/service/auth" '@{CredSSP="true"}'
cmd.exe /c winrm set "winrm/config/listener?Address=*+Transport=HTTPS" "@{Port=`"5986`";Hostname=`"packer`";CertificateThumbprint=`"$($Cert.Thumbprint)`"}"
cmd.exe /c netsh advfirewall firewall set rule group="remote administration" new enable=yes
cmd.exe /c netsh advfirewall firewall add rule name="Port 5986" dir=in action=allow protocol=TCP localport=5986 profile=any
cmd.exe /c net stop winrm
cmd.exe /c sc config winrm start= auto
cmd.exe /c net start winrm

</powershell>
EOF
}

hcp_packer_registry {
  bucket_name = "native-sbom"
  description = "Windows SBOM test using a legacy execute_command style."
}

build {
  name    = "native-sbom-packer-windows-legacy-exec"
  sources = ["source.amazon-ebs.windows"]

  provisioner "powershell" {
    inline = ["Write-Output 'packer works on Windows'"]
  }

  provisioner "hcp-sbom" {
    auto_generate = true
    scan_path     = "C:\\Windows\\Temp"
    destination   = "sbom.json"
    sbom_name     = "windows-legacy-exec"
    scanner_args  = ["-o", "cyclonedx-json"]

    # Legacy style command that omits 'sbom-generate'.
    execute_command = "{{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}"
  }
}
