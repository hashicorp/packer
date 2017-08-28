---
description: |
    The posix restart provisioner restarts a Unix/Linux machine and waits for it
    to come back up.
layout: docs
page_title: Posix Restart Provisioner
...

# Posix Restart Provisioner

Type: `posix-restart`

The posix restart provisioner initiates a reboot on a Unix/Linux machine and
waits for the machine to come back online.

The Unix/Linux provisioning process rarely requires reboots, this
provisioner helps to ease that process if needed

Packer expects the machine to be ready to continue provisioning after it
reboots. Packer detects that the reboot has completed by making an SSH call,
 not by ACPI functions, so the operating system must be completely booted in order to continue.

## Basic Example

The example below is fully functional.

``` {.javascript}
{
  "type": "posix-restart"
}
```

## Configuration Reference

The reference of available configuration options is listed below.

Optional parameters:

-   `restart_commands` (array) - The commands to execute to initiate the
    restart. By default this is the lines of a script that try to detect service
    manager used and do the corresponding command
    `nohup sh -c 'systemctl stop sshd && reboot`,
    `nohup sh -c 'service sshd stop && shutdown -r now` or
    `nohup sh -c '/etc/init.d/sshd stop && shutdown -r now`.
    A key action of this is to stop SSH daemon so that Packer can
    detect it is rebooting.

-   `restart_check_command` (string) - A command to execute to check if the
    restart succeeded. This will be done in a loop.

-   `restart_timeout` (string) - The timeout to wait for the restart. By
    default this is 5 minutes. Example value: `5m`. If you are installing
    updates or have a lot of startup services, you will probably need to
    increase this duration.

-   `execute_command` (string) - The command to use to execute the script. By
    default this is `chmod +x {{ .Path }}; {{ .Vars }} {{ .Path }}`. The value
    of this is treated as [configuration
    template](/docs/templates/configuration-templates.html). There are two
    available variables: `Path`, which is the path to the script to run, and
    `Vars`, which is the list of `environment_vars`, if configured.
