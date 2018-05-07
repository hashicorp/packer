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

Exactly *one* of the following is required:

-   `command` (string) - This is a single command to execute. It will be written
    to a temporary file and run using the `execute_command` call below.

-   `inline` (array of strings) - This is an array of commands to execute. The
    commands are concatenated by newlines and turned into a single file, so they
    are all executed within the same context. This allows you to change
    directories in one command and use something in the directory in the next
    and so on. Inline scripts are the easiest way to pull off simple tasks
    within the machine.

-   `script` (string) - The path to a script to execute. This path can be
    absolute or relative. If it is relative, it is relative to the working
    directory when Packer is executed.

-   `scripts` (array of strings) - An array of scripts to execute. The scripts
    will be executed in the order specified. Each script is executed in
    isolation, so state such as variables from one script won't carry on to the
    next.

Optional parameters:

-   `execute_command` (array of strings) - The command to use to execute
    the script. By default this is `["/bin/sh", "-c", "{{.Command}}"]`. The value
    is an array of arguments executed directly by the OS. The value of this is
    treated as [configuration
    template](/docs/templates/engine.html). The only available
    variable is `Command` which is the command to execute.

-   `environment_vars` (array of strings) - An array of key/value pairs to
    inject prior to the `execute_command`. The format should be `key=value`.
    Packer injects some environmental variables by default into the environment,
    as well, which are covered in the section below.

-   `execute_command` (array of strings) - The command used to execute the script.
    By default this is `["/bin/sh", "-c", "{{.Vars}}, "{{.Script}}"]`
    on unix and `["cmd", "/c", "{{.Vars}}", "{{.Script}}"]` on windows.
    This is treated as a [template engine](/docs/templates/engine.html).
    There are two available variables: `Script`, which is the path to the script
    to run, and `Vars`, which is the list of `environment_vars`, if configured
    If you choose to set this option, make sure that the first element in the
    array is the shell program you want to use (for example, "sh" or
    "/usr/local/bin/zsh" or even "powershell.exe" although anything other than
    a flavor of the shell command language is not explicitly supported and may
    be broken by assumptions made within Packer), and a later element in the
    array must be `{{.Script}}`.

    For backwards compatability, {{.Command}} is also available to use in
    `execute_command` but it is decoded the same way as {{.Script}}. We
    recommend using {{.Script}} for the sake of clarity, as even when you set
    only a single `command` to run, Packer writes it to a temporary file and
    then runs it as a script.

-   `inline_shebang` (string) - The
    [shebang](http://en.wikipedia.org/wiki/Shebang_%28Unix%29) value to use when
    running commands specified by `inline`. By default, this is `/bin/sh -e`. If
    you're not using `inline`, then this configuration has no effect.
    **Important:** If you customize this, be sure to include something like the
    `-e` flag, otherwise individual steps failing won't fail the provisioner.

-  `use_linux_pathing` (bool) - This is only relevant to windows hosts. If you
   are running Packer in a Windows environment with the Windows Subsystem for
   Linux feature enabled, and would like to invoke a bash script rather than
   invoking a Cmd script, you'll need to set this flag to true; it tells Packer
   to use the linux subsystem path for your script rather than the Windows path.
   (e.g. /mnt/c/path/to/your/file instead of C:/path/to/your/file). Please see
   the example below for more guidance on how to use this feature. If you are
   not on a Windows host, or you do not intend to use the shell-local
   provisioner to run a bash script, please ignore this option.

## Execute Command

To many new users, the `execute_command` is puzzling. However, it provides an
important function: customization of how the command is executed. The most
common use case for this is dealing with **sudo password prompts**. You may also
need to customize this if you use a non-POSIX shell, such as `tcsh` on FreeBSD.

### The Windows Linux Subsystem

The shell-local provisioner was designed with the idea of allowing you to run
commands in your local operating system's native shell. For Windows, we've
assumed in our defaults that this is Cmd. However, it is possible to run a
bash script as part of the Windows Linux Subsystem from the shell-local
provisioner, by modifying the `execute_command` and the `use_linux_pathing`
options in the provisioner config.

The example below is a fully functional test config.

One limitation of this offering is that "inline" and "command" options are not
available to you; please limit yourself to using the "script" or "scripts"
options instead.

Please note that the WSL is a beta feature, and this tool is not guaranteed to
work as you expect it to.

```
{
    "builders": [
      {
        "type":         "null",
        "communicator": "none"
      }
    ],
    "provisioners": [
      {
          "type": "shell-local",
          "environment_vars": ["PROVISIONERTEST=ProvisionerTest1"],
          "execute_command": ["bash", "-c", "{{.Vars}} {{.Script}}"],
          "use_linux_pathing": true,
          "scripts": ["C:/Users/me/scripts/example_bash.sh"]
      },
      {
          "type": "shell-local",
          "environment_vars": ["PROVISIONERTEST=ProvisionerTest2"],
          "execute_command": ["bash", "-c", "{{.Vars}} {{.Script}}"],
          "use_linux_pathing": true,
          "script": "C:/Users/me/scripts/example_bash.sh"
      }
  ]
}
```

## Default Environmental Variables

In addition to being able to specify custom environmental variables using the
`environment_vars` configuration, the provisioner automatically defines certain
commonly useful environmental variables:

-   `PACKER_BUILD_NAME` is set to the name of the build that Packer is running.
    This is most useful when Packer is making multiple builds and you want to
    distinguish them slightly from a common provisioning script.

-   `PACKER_BUILDER_TYPE` is the type of the builder that was used to create the
    machine that the script is running on. This is useful if you want to run
    only certain parts of the script on systems built with certain builders.

## Safely Writing A Script

Whether you use the `inline` option, or pass it a direct `script` or `scripts`,
it is important to understand a few things about how the shell-local
provisioner works to run it safely and easily. This understanding will save
you much time in the process.

### Once Per Builder

The `shell-local` script(s) you pass are run once per builder. That means that
if you have an `amazon-ebs` builder and a `docker` builder, your script will be
run twice. If you have 3 builders, it will run 3 times, once for each builder.

### Always Exit Intentionally

If any provisioner fails, the `packer build` stops and all interim artifacts
are cleaned up.

For a shell script, that means the script **must** exit with a zero code. You
*must* be extra careful to `exit 0` when necessary.
