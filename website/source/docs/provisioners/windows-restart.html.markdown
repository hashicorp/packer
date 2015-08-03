---
description: |
    The Windows restart provisioner restarts a Windows machine and waits for it to
    come back up.
layout: docs
page_title: Windows Restart Provisioner
...

# Windows Restart Provisioner

Type: `windows-restart`

The Windows restart provisioner initiates a reboot on a Windows machine and
waits for the machine to come back online.

The Windows provisioning process often requires multiple reboots, and this
provisioner helps to ease that process.

## Basic Example

The example below is fully functional.

``` {.javascript}
{
  "type": "windows-restart"
}
```

## Configuration Reference

The reference of available configuration options is listed below.

Optional parameters:

-   `restart_command` (string) - The command to execute to initiate the restart.
    By default this is `shutdown /r /c "packer restart" /t 5 && net stop winrm`.
    A key action of this is to stop WinRM so that Packer can detect it
    is rebooting.

-   `restart_check_command` (string) - A command to execute to check if the
    restart succeeded. This will be done in a loop.

-   `restart_timeout` (string) - The timeout to wait for the restart. By default
    this is 5 minutes. Example value: "5m"
