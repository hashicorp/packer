---
description: |
    The shell Packer provisioner provisions machines built by Packer using shell
    scripts. Shell provisioning is the easiest way to get software installed and
    configured on a machine.
layout: docs
page_title: PowerShell Provisioner
...

# PowerShell Provisioner

Type: `powershell`

The PowerShell Packer provisioner runs PowerShell scripts on Windows machines.
It assumes that the communicator in use is WinRM.

## Basic Example

The example below is fully functional.

``` {.javascript}
{
  "type": "powershell",
  "inline": ["dir c:\\"]
}
```

## Configuration Reference

The reference of available configuration options is listed below. The only
required element is either "inline" or "script". Every other option is optional.

Exactly *one* of the following is required:

-   `inline` (array of strings) - This is an array of commands to execute. The
    commands are concatenated by newlines and turned into a single file, so they
    are all executed within the same context. This allows you to change
    directories in one command and use something in the directory in the next
    and so on. Inline scripts are the easiest way to pull off simple tasks
    within the machine.

-   `script` (string) - The path to a script to upload and execute in
    the machine. This path can be absolute or relative. If it is relative, it is
    relative to the working directory when Packer is executed.

-   `scripts` (array of strings) - An array of scripts to execute. The scripts
    will be uploaded and executed in the order specified. Each script is
    executed in isolation, so state such as variables from one script won't
    carry on to the next.

Optional parameters:

-   `binary` (boolean) - If true, specifies that the script(s) are binary files,
    and Packer should therefore not convert Windows line endings to Unix line
    endings (if there are any). By default this is false.

-   `environment_vars` (array of strings) - An array of key/value pairs to
    inject prior to the execute\_command. The format should be `key=value`.
    Packer injects some environmental variables by default into the environment,
    as well, which are covered in the section below.

-   `execute_command` (string) - The command to use to execute the script. By
    default this is `powershell "& { {{.Vars}}{{.Path}}; exit $LastExitCode}"`.
    The value of this is treated as [configuration
    template](/docs/templates/configuration-templates.html). There are two
    available variables: `Path`, which is the path to the script to run, and
    `Vars`, which is the list of `environment_vars`, if configured.

-   `elevated_user` and `elevated_password` (string) - If specified, the
    PowerShell script will be run with elevated privileges using the given
    Windows user.

-   `remote_path` (string) - The path where the script will be uploaded to in
    the machine. This defaults to "c:/Windows/Temp/script.ps1". This value must be a
    writable location and any parent directories must already exist.

-   `start_retry_timeout` (string) - The amount of time to attempt to *start*
    the remote process. By default this is "5m" or 5 minutes. This setting
    exists in order to deal with times when SSH may restart, such as a
    system reboot. Set this to a higher value if reboots take a longer amount
    of time.

-   `valid_exit_codes` (list of ints) - Valid exit codes for the script. By
    default this is just 0.
