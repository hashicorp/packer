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

## Safely Writing A Script

Whether you use the `inline` option, or pass it a direct `script` or `scripts`, it is important to understand a few things about how the shell-local post-processor works to run it safely and easily. This understanding will save you much time in the process.

### Once Per Artifact

The `shell-local` script(s) you pass are run once per artifact output file. That means that if your builder results in 1 output file, your script will be run once. If it results in 3 output files, it will run 3 times, once for each file.

For example, the virtualbox builders, when configured to provide an `ovf` output format (the default), will provide **two** output files:

* The actual disk itself, in `.vmdk` format
* The appliance description file, in `.ovf` format

Each time each shell-local script is run, it is passed the path to the artifact file, relative to the directory in which packer is run, as the first argument to the script. 

Let's take a simple example. You want to run a post-processor that records the name of every artifact created to `/tmp/artifacts`. (Why? I don't know. For fun.)

Your post-processor should look like this:


``` {.javascript}
{
  "type": "shell-local",
  "inline": [
    "echo \$1 >> /tmp/artifacts"
  ]
}
```

The result of the above will be an output line for each artifact.

The net effect of this is that if you want to post-process only some files, **you must test** `$1` to see if it is the file you want.

Here is an example script that converts the `.vmdk` artifact of a virtualbox build to a raw img, suitable for converting to a USB.


``` {.bash}
    #!/bin/bash -e

    [[ "$1" == *.vmdk ]] && vboxmanage clonemedium disk $1 --format raw output_file.img

```

### Always Exit Intentionally

If any post-processor fails, the `packer build` stops and all interim artifacts are cleaned up.

For a shell script, that means the script **must** exit with a zero code. You *must* be extra careful to `exit 0` when necessary. Using our above conversion script example, if the current artifact is *not* a `.vmdk` file, the test `[[ "$1" == *.vmdk ]]` will fail. Since that is the last command in the script, the script will exit with a non-zero code, the post-processor will fail, the build will fail, and you will have to start over.

Of course, we didn't mean that! We just meant:

* If a `.vmdk` file, convert, and that is OK
* If not a `.vmdk` file, ignore, and that is OK

To make it work correctly, use the following instead:

``` {.bash}
    #!/bin/bash -e

    [[ "$1" == *.vmdk ]] && vboxmanage clonemedium disk $1 --format raw output_file.img

    # always exit 0 unless a command actually fails
    exit 0

````

