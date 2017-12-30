---
description: |
    shell-local will run a shell script of your choosing on the machine where Packer
    is being run - in other words, it shell-local will run the shell script on your
    build server, or your desktop, etc., rather than the remote/guest machine being
    provisioned by Packer.
layout: docs
page_title: 'Shell (Local) - Provisioners'
sidebar_current: 'docs-provisioners-shell-local'
---

# Local Shell Provisioner

Type: `shell-local`

shell-local will run a shell script of your choosing on the machine where Packer
is being run - in other words, it shell-local will run the shell script on your
build server, or your desktop, etc., rather than the remote/guest machine being
provisioned by Packer.

The [remote shell](/docs/provisioners/shell.html) provisioner executes
shell scripts on a remote machine.

## Basic Example

The example below is fully functional.

``` json
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
    template](/docs/templates/engine.html). The only available
    variable is `Command` which is the command to execute.
