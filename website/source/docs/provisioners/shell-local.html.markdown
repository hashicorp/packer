---
description: |
    The shell Packer provisioner provisions machines built by Packer using shell
    scripts. Shell provisioning is the easiest way to get software installed and
    configured on a machine.
layout: docs
page_title: Local Shell Provisioner
...

# Local Shell Provisioner

Type: `shell-local`

The local shell provisioner executes a local shell script on the machine running
Packer. The [remote shell](/docs/provisioners/shell.html) provisioner executes
shell scripts on a remote machine.

## Basic Example

The example below is fully functional.

``` {.javascript}
{
  "type": "shell-local",
  "command": "echo foo"
}
```

## Configuration Reference

The reference of available configuration options is listed below. The only
required element is "command".

Required:

-   `command` (string) - The command to execute. This will be executed within
    the context of a shell as specified by `execute_command`.

Optional parameters:

-   `execute_command` (array of strings) - The command to use to execute
    the script. By default this is `["/bin/sh", "-c", "{{.Command}}"]`. The value
    is an array of arguments executed directly by the OS. The value of this is
    treated as [configuration
    template](/docs/templates/configuration-templates.html). The only available
    variable is `Command` which is the command to execute.
