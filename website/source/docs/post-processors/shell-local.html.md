---
description: |
    The shell-local Packer post processor enables users to do some post processing after artifacts have been built.
layout: docs
page_title: Local Shell Post Processor
...

# Local Shell Post Processor

Type: `shell-local`

The local shell post processor executes scripts locally during the post processing stage. Shell local provides an easy
way to automate executing some task with the packer outputs.

## Basic example

The example below is fully functional.

``` {.javascript}
{
  "type": "shell-local",
  "inline": ["echo foo"]
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

-   `environment_vars` (array of strings) - An array of key/value pairs to
    inject prior to the execute\_command. The format should be `key=value`.
    Packer injects some environmental variables by default into the environment,
    as well, which are covered in the section below.

-   `execute_command` (string) - The command to use to execute the script. By
    default this is `chmod +x {{.Script}}; {{.Vars}} {{.Script}} {{.Artifact}}`.
    The value of this is treated as [configuration template](/docs/templates/configuration-templates.html).
    There are three available variables: `Script`, which is the path to the script
    to run, `Vars`, which is the list of `environment_vars`, if configured and
    `Artifact`, which is path to artifact file.

-   `inline_shebang` (string) - The
    [shebang](http://en.wikipedia.org/wiki/Shebang_%28Unix%29) value to use when
    running commands specified by `inline`. By default, this is `/bin/sh -e`. If
    you're not using `inline`, then this configuration has no effect.
    **Important:** If you customize this, be sure to include something like the
    `-e` flag, otherwise individual steps failing won't fail the provisioner.

## Execute Command Example

To many new users, the `execute_command` is puzzling. However, it provides an
important function: customization of how the command is executed. The most
common use case for this is dealing with **sudo password prompts**. You may also
need to customize this if you use a non-POSIX shell, such as `tcsh` on FreeBSD.

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
