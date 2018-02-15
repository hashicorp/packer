---
description: |
    The windows-shell Packer provisioner runs commands on Windows using the cmd
    shell.
layout: docs
page_title: 'Windows Shell - Provisioners'
sidebar_current: 'docs-provisioners-windows-shell'
---

# Windows Shell Provisioner

Type: `windows-shell`

The windows-shell Packer provisioner runs commands on a Windows machine using
`cmd`. It assumes it is running over WinRM.

## Basic Example

The example below is fully functional.

``` json
{
  "type": "windows-shell",
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
    default this is `{{ .Vars }}"{{ .Path }}"`. The value of this is treated as
    [template engine](/docs/templates/engine.html).
    There are two available variables: `Path`, which is the path to the script
    to run, and `Vars`, which is the list of `environment_vars`, if configured.

-   `remote_path` (string) - The path where the script will be uploaded to in
    the machine. This defaults to "c:/Windows/Temp/script.bat". This value must be a
    writable location and any parent directories must already exist.

-   `start_retry_timeout` (string) - The amount of time to attempt to *start*
    the remote process. By default this is "5m" or 5 minutes. This setting
    exists in order to deal with times when SSH may restart, such as a
    system reboot. Set this to a higher value if reboots take a longer amount
    of time.

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

-   `PACKER_HTTP_ADDR` If using a builder that provides an http server for file
    transfer (such as hyperv, parallels, qemu, virtualbox, and vmware), this
    will be set to the address. You can use this address in your provisioner to
    download large files over http. This may be useful if you're experiencing
    slower speeds using the default file provisioner. A file provisioner using
    the `winrm` communicator may experience these types of difficulties.
