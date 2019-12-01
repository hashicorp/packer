---
description: |
    Communicators are the mechanism Packer uses to upload files, execute scripts,
    etc. with the machine being created.
layout: docs
page_title: 'Communicators - WinRM'
sidebar_current: 'docs-communicators-winrm'
---

# WinRM Communicator

The WinRM communicator connects to the host via WinRM if it is enabled and configured on the remote host. WinRM is not enabled by default on many base images; for an example of how to configure WinRM before provisioning in AWS, see [this page](https://www.packer.io/intro/getting-started/build-image.html#a-windows-example).

The WinRM communicator has the following options.

-   `winrm_host` (string) - The address for WinRM to connect to.

    NOTE: If using an Amazon EBS builder, you can specify the interface WinRM
    connects to via
    [`ssh_interface`](https://www.packer.io/docs/builders/amazon-ebs.html#ssh_interface)

-   `winrm_insecure` (boolean) - If `true`, do not check server certificate
    chain and host name.

-   `winrm_password` (string) - The password to use to connect to WinRM.

-   `winrm_port` (number) - The WinRM port to connect to. This defaults to
    `5985` for plain unencrypted connection and `5986` for SSL when
    `winrm_use_ssl` is set to true.

-   `winrm_timeout` (string) - The amount of time to wait for WinRM to become
    available. This defaults to `30m` since setting up a Windows machine
    generally takes a long time.

-   `winrm_use_ntlm` (boolean) - If `true`, NTLMv2 authentication (with session
    security) will be used for WinRM, rather than default (basic
    authentication), removing the requirement for basic authentication to be
    enabled within the target guest. Further reading for remote connection
    authentication can be found
    [here](https://msdn.microsoft.com/en-us/library/aa384295(v=vs.85).aspx).

-   `winrm_use_ssl` (boolean) - If `true`, use HTTPS for WinRM.

-   `winrm_username` (string) - The username to use to connect to WinRM.

## Configuring WinRM as part of an Autounattend File

You can add a batch file to your autounattend that contains the commands for
configuring winrm. Depending on your winrm setup, this could be a complex batch
file, or a very simple one.

``` xml
<FirstLogonCommands>
  ...
  <SynchronousCommand wcm:action="add">
      <CommandLine>cmd.exe /c a:\winrmConfig.bat</CommandLine>
      <Description>Configure WinRM</Description>
      <Order>3</Order>
      <RequiresUserInput>true</RequiresUserInput>
  </SynchronousCommand>
  ...
</FirstLogonCommands>
```

The winrmConfig.bat referenced above can be as simple as

```
rem basic config for winrm
cmd.exe /c winrm quickconfig -q

rem allow unencrypted traffic, and configure auth to use basic username/password auth
cmd.exe /c winrm set winrm/config/service @{AllowUnencrypted="true"}
cmd.exe /c winrm set winrm/config/service/auth @{Basic="true"}

rem update firewall rules to open the right port and to allow remote administration
cmd.exe /c netsh advfirewall firewall set rule group="remote administration" new enable=yes

rem restart winrm
cmd.exe /c net stop winrm
cmd.exe /c net start winrm
```

This batch file will only work for http connections, not https, but will enable
you to connect using only the username and password created earlier in the
Autounattend file. The above batchfile will allow you to connect using a very
simple Packer config:

```json
        ...
        "communicator": "winrm",
        "winrm_username": "packeruser",
        "winrm_password": "SecretPassword"
        ...
```

If you want to set winRM up for https, things will be a bit more complicated.
We'll explore
