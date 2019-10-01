---
description: |
    Communicators are the mechanism Packer uses to upload files, execute scripts,
    etc. with the machine being created.
layout: docs
page_title: 'Communicators - Templates'
sidebar_current: 'docs-templates-communicators'
---

# Template Communicators

Communicators are the mechanism Packer uses to upload files, execute scripts,
etc. with the machine being created.

Communicators are configured within the
[builder](/docs/templates/builders.html) section. Packer currently supports
three kinds of communicators:

-   `none` - No communicator will be used. If this is set, most provisioners
    also can't be used.

-   `ssh` - An SSH connection will be established to the machine. This is
    usually the default.

-   `winrm` - A WinRM connection will be established.

In addition to the above, some builders have custom communicators they can use.
For example, the Docker builder has a "docker" communicator that uses
`docker exec` and `docker cp` to execute scripts and copy files.

## Using a Communicator

By default, the SSH communicator is usually used. Additional configuration may
not even be necessary, since some builders such as Amazon automatically
configure everything.

However, to specify a communicator, you set the `communicator` key within a
build. Multiple builds can have different communicators. Example:

``` json
{
  "builders": [
    {
      "type": "amazon-ebs",
      "communicator": "ssh"
    }
  ]
}
```

After specifying the `communicator`, you can specify a number of other
configuration parameters for that communicator. These are documented below.

## WinRM Communicator

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

## Pausing Before Connecting
We recommend that you enable SSH or WinRM as the very last step in your
guest's bootstrap script, but sometimes you may have a race condition where
you need Packer to wait before attempting to connect to your guest.

If you end up in this situation, you can use the template option
`pause_before_connecting`. By default, there is no pause. For example:

```
{
  "communicator": "ssh",
  "ssh_username": "myuser",
  "pause_before_connecting": "10m"
}
```

In this example, Packer will check whether it can connect, as normal. But once
a connection attempt is successful, it will disconnect and then wait 10 minutes
before connecting to the guest and beginning provisioning.


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