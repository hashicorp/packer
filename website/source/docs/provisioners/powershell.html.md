---
description: |
    The PowerShell Packer provisioner runs PowerShell scripts on Windows
    machines.
    It assumes that the communicator in use is WinRM.
layout: docs
page_title: 'PowerShell - Provisioners'
sidebar_current: 'docs-provisioners-powershell'
---

# PowerShell Provisioner

Type: `powershell`

The PowerShell Packer provisioner runs PowerShell scripts on Windows machines.
It assumes that the communicator in use is WinRM. However, the provisioner
can work equally well (with a few caveats) when combined with the SSH
communicator. See the [section
below](/docs/provisioners/powershell.html#combining-the-powershell-provisioner-with-the-ssh-communicator)
for details.

## Basic Example

The example below is fully functional.

``` json
{
  "type": "powershell",
  "inline": ["dir c:\\"]
}
```

## Configuration Reference

The reference of available configuration options is listed below. The only
required element is either "inline" or "script". Every other option is
optional.

Exactly *one* of the following is required:

-   `inline` (array of strings) - This is an array of commands to execute. The
    commands are concatenated by newlines and turned into a single file, so
    they are all executed within the same context. This allows you to change
    directories in one command and use something in the directory in the next
    and so on. Inline scripts are the easiest way to pull off simple tasks
    within the machine.

-   `script` (string) - The path to a script to upload and execute in
    the machine. This path can be absolute or relative. If it is relative, it
    is relative to the working directory when Packer is executed.

-   `scripts` (array of strings) - An array of scripts to execute. The scripts
    will be uploaded and executed in the order specified. Each script is
    executed in isolation, so state such as variables from one script won't
    carry on to the next.

Optional parameters:

-   `binary` (boolean) - If true, specifies that the script(s) are binary
    files, and Packer should therefore not convert Windows line endings to Unix
    line endings (if there are any). By default this is false.

-   `elevated_execute_command` (string) - The command to use to execute the
    elevated script. By default this is as follows:

    ``` powershell
    powershell -executionpolicy bypass "& { if (Test-Path variable:global:ProgressPreference){$ProgressPreference='SilentlyContinue'};. {{.Vars}}; &'{{.Path}}'; exit $LastExitCode }"
    ```

    The value of this is treated as [configuration
    template](/docs/templates/engine.html). There are two
    available variables: `Path`, which is the path to the script to run, and
    `Vars`, which is the location of a temp file containing the list of
    `environment_vars`, if configured.

-   `environment_vars` (array of strings) - An array of key/value pairs to
    inject prior to the execute\_command. The format should be `key=value`.
    Packer injects some environmental variables by default into the
    environment, as well, which are covered in the section below.
    If you are running on AWS, Azure or Google Compute and would like to access the generated
    password that Packer uses to connect to the instance via
    WinRM, you can use the template variable `{{.WinRMPassword}}` to set this
    as an environment variable. For example:

    ```json
      {
        "type": "powershell",
        "environment_vars": "WINRMPASS={{.WinRMPassword}}",
        "inline": ["Write-Host \"Automatically generated aws password is: $Env:WINRMPASS\""]
      },
    ```

-   `execute_command` (string) - The command to use to execute the script. By
    default this is as follows:

    ``` powershell
    powershell -executionpolicy bypass "& { if (Test-Path variable:global:ProgressPreference){$ProgressPreference='SilentlyContinue'};. {{.Vars}}; &'{{.Path}}'; exit $LastExitCode }"
    ```

    The value of this is treated as [configuration
    template](/docs/templates/engine.html). There are two
    available variables: `Path`, which is the path to the script to run, and
    `Vars`, which is the location of a temp file containing the list of
    `environment_vars`. The value of both `Path` and `Vars` can be
    manually configured by setting the values for `remote_path` and
    `remote_env_var_path` respectively.

-   `elevated_user` and `elevated_password` (string) - If specified, the
    PowerShell script will be run with elevated privileges using the given
    Windows user. If you are running a build on AWS, Azure or Google Compute and would like to run using
    the generated password that Packer uses to connect to the instance via 
    WinRM, you may do so by using the template variable {{.WinRMPassword}}.
    For example:

    ``` json
    "elevated_user": "Administrator",
    "elevated_password": "{{.WinRMPassword}}",
    ```

-   `remote_path` (string) - The path where the PowerShell script will be
    uploaded to within the target build machine. This defaults to
    `C:/Windows/Temp/script-UUID.ps1` where UUID is replaced with a
    dynamically generated string that uniquely identifies the script.

    This setting allows users to override the default upload location. The
    value must be a writable location and any parent directories must
    already exist.

-   `remote_env_var_path` (string) - Environment variables required within
    the remote environment are uploaded within a PowerShell script and then
    enabled by 'dot sourcing' the script immediately prior to execution of
    the main command or script.

    The path the environment variables script will be uploaded to defaults to
    `C:/Windows/Temp/packer-ps-env-vars-UUID.ps1` where UUID is replaced
    with a dynamically generated string that uniquely identifies the
    script.

    This setting allows users to override the location the environment
    variable script is uploaded to. The value must be a writable location
    and any parent directories must already exist.

-   `start_retry_timeout` (string) - The amount of time to attempt to *start*
    the remote process. By default this is "5m" or 5 minutes. This setting
    exists in order to deal with times when SSH may restart, such as a
    system reboot. Set this to a higher value if reboots take a longer amount
    of time.

-   `valid_exit_codes` (list of ints) - Valid exit codes for the script. By
    default this is just 0.

## Default Environmental Variables

In addition to being able to specify custom environmental variables using the
`environment_vars` configuration, the provisioner automatically defines certain
commonly useful environmental variables:

-   `PACKER_BUILD_NAME` is set to the
    [name of the build](/docs/templates/builders.html#named-builds) that Packer is running.
    This is most useful when Packer is making multiple builds and you want to
    distinguish them slightly from a common provisioning script.

-   `PACKER_BUILDER_TYPE` is the type of the builder that was used to create
    the machine that the script is running on. This is useful if you want to
    run only certain parts of the script on systems built with certain
    builders.

-   `PACKER_HTTP_ADDR` If using a builder that provides an http server for file
    transfer (such as hyperv, parallels, qemu, virtualbox, and vmware), this
    will be set to the address. You can use this address in your provisioner to
    download large files over http. This may be useful if you're experiencing
    slower speeds using the default file provisioner. A file provisioner using
    the `winrm` communicator may experience these types of difficulties.

## Combining the PowerShell Provisioner with the SSH Communicator

The good news first. If you are using the
[Microsoft port of OpenSSH](https://github.com/PowerShell/Win32-OpenSSH/wiki)
then the provisioner should just work as expected - no extra configuration
effort is required.

Now the caveats. If you are using an alternative configuration, and your SSH
connection lands you in a *nix shell on the remote host, then you will most
likely need to manually set the `execute_command`; The default
`execute_command` used by Packer will not work for you.
When configuring the command you will need to ensure that any dollar signs
or other characters that may be incorrectly interpreted by the remote shell
are escaped accordingly.

The following example shows how the standard `execute_command` can be
reconfigured to work on a remote system with
[Cygwin/OpenSSH](https://cygwin.com/) installed.
The `execute_command` has each dollar sign backslash escaped so that it is
not interpreted by the remote Bash shell - Bash being the default shell for
Cygwin environments.

```json
  "provisioners": [
    {
      "type": "powershell",
      "execute_command": "powershell -executionpolicy bypass \"& { if (Test-Path variable:global:ProgressPreference){\\$ProgressPreference='SilentlyContinue'};. {{.Vars}}; &'{{.Path}}'; exit \\$LastExitCode }\"",
      "inline": [
        "Write-Host \"Hello from PowerShell\"",
      ]
    }
  ]
```


## Packer's Handling of Characters Special to PowerShell

The escape character in PowerShell is the `backtick`, also sometimes
referred to as the `grave accent`. When, and when not, to escape characters
special to PowerShell is probably best demonstrated with a series of examples.

### When To Escape...

Users need to deal with escaping characters special to PowerShell when they
appear *directly* in commands used in the `inline` PowerShell provisioner and
when they appear *directly* in the users own scripts.
Note that where double quotes appear within double quotes, the addition of
a backslash escape is required for the JSON template to be parsed correctly.

``` json
  "provisioners": [
    {
      "type": "powershell",
      "inline": [
          "Write-Host \"A literal dollar `$ must be escaped\"",
          "Write-Host \"A literal backtick `` must be escaped\"",
          "Write-Host \"Here `\"double quotes`\" must be escaped\"",
          "Write-Host \"Here `'single quotes`' don`'t really need to be\"",
          "Write-Host \"escaped... but it doesn`'t hurt to do so.\"",
      ]
    },
```

The above snippet should result in the following output on the Packer console:

```
==> amazon-ebs: Provisioning with Powershell...
==> amazon-ebs: Provisioning with powershell script: /var/folders/15/d0f7gdg13rnd1cxp7tgmr55c0000gn/T/packer-powershell-provisioner508190439
    amazon-ebs: A literal dollar $ must be escaped
    amazon-ebs: A literal backtick ` must be escaped
    amazon-ebs: Here "double quotes" must be escaped
    amazon-ebs: Here 'single quotes' don't really need to be
    amazon-ebs: escaped... but it doesn't hurt to do so.
```

### When Not To Escape...

Special characters appearing in user environment variable values and in the
`elevated_user` and `elevated_password` fields will be automatically
dealt with for the user. There is no need to use escapes in these instances.

``` json
{
  "variables": {
    "psvar": "My$tring"
  },
  ...
  "provisioners": [
    {
      "type": "powershell",
      "elevated_user": "Administrator",
      "elevated_password": "Super$3cr3t!",
      "inline": "Write-Output \"The dollar in the elevated_password is interpreted correctly\""
    },
    {
      "type": "powershell",
      "environment_vars": [
        "VAR1=A$Dollar",
        "VAR2=A`Backtick",
        "VAR3=A'SingleQuote",
        "VAR4=A\"DoubleQuote",
        "VAR5={{user `psvar`}}"
      ],
      "inline": [
        "Write-Output \"In the following examples the special character is interpreted correctly:\"",
        "Write-Output \"The dollar in VAR1:                            $Env:VAR1\"",
        "Write-Output \"The backtick in VAR2:                          $Env:VAR2\"",
        "Write-Output \"The single quote in VAR3:                      $Env:VAR3\"",
        "Write-Output \"The double quote in VAR4:                      $Env:VAR4\"",
        "Write-Output \"The dollar in VAR5 (expanded from a user var): $Env:VAR5\""
      ]
    }
  ]
  ...
}
```

The above snippet should result in the following output on the Packer console:

```
==> amazon-ebs: Provisioning with Powershell...
==> amazon-ebs: Provisioning with powershell script: /var/folders/15/d0f7gdg13rnd1cxp7tgmr55c0000gn/T/packer-powershell-provisioner961728919
    amazon-ebs: The dollar in the elevated_password is interpreted correctly
==> amazon-ebs: Provisioning with Powershell...
==> amazon-ebs: Provisioning with powershell script: /var/folders/15/d0f7gdg13rnd1cxp7tgmr55c0000gn/T/packer-powershell-provisioner142826554
    amazon-ebs: In the following examples the special character is interpreted correctly:
    amazon-ebs: The dollar in VAR1:                            A$Dollar
    amazon-ebs: The backtick in VAR2:                          A`Backtick
    amazon-ebs: The single quote in VAR3:                      A'SingleQuote
    amazon-ebs: The double quote in VAR4:                      A"DoubleQuote
    amazon-ebs: The dollar in VAR5 (expanded from a user var): My$tring
```
