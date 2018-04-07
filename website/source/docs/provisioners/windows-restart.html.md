---
description: |
    The Windows restart provisioner restarts a Windows machine and waits for it to
    come back up.
layout: docs
page_title: 'Windows Restart - Provisioners'
sidebar_current: 'docs-provisioners-windows-restart'
---

# Windows Restart Provisioner

Type: `windows-restart`

The Windows restart provisioner initiates a reboot on a Windows machine and
waits for the machine to come back online.

The Windows provisioning process often requires multiple reboots, and this
provisioner helps to ease that process.

Packer expects the machine to be ready to continue provisioning after it
reboots. Packer detects that the reboot has completed by making an RPC call
through the Windows Remote Management (WinRM) service, not by ACPI functions, so Windows must be completely booted in order to continue.

## Basic Example

The example below is fully functional.

``` json
{
  "type": "windows-restart"
}
```

## Configuration Reference

The reference of available configuration options is listed below.

Optional parameters:

-   `restart_command` (string) - The command to execute to initiate the
    restart. By default this is `shutdown /r /f /t 0 /c "packer restart"`.

-   `restart_check_command` (string) - A command to execute to check if the
    restart succeeded. This will be done in a loop. Example usage:

``` json
    {
      "type": "windows-restart",
      "restart_check_command": "powershell -command \"& {Write-Output 'restarted.'}\""
    },
```

-   `restart_timeout` (string) - The timeout to wait for the restart. By
    default this is 5 minutes. Example value: `5m`. If you are installing
    updates or have a lot of startup services, you will probably need to
    increase this duration.
