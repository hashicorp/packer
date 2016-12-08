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

-   `script` (string) - The path to a script to execute. This path can be
    absolute or relative. If it is relative, it is relative to the working
    directory when Packer is executed.

-   `scripts` (array of strings) - An array of scripts to execute. The scripts
    will be executed in the order specified. Each script is executed in
    isolation, so state such as variables from one script won't carry on to the
    next.

Optional parameters:

-   `environment_vars` (array of strings) - An array of key/value pairs to
    inject prior to the execute\_command. The format should be `key=value`.
    Packer injects some environmental variables by default into the environment,
    as well, which are covered in the section below.

-   `execute_command` (string) - The command to use to execute the script. By
    default this is `chmod +x "{{.Script}}"; {{.Vars}} "{{.Script}}"`.
    The value of this is treated as [configuration template](/docs/templates/configuration-templates.html).
    There are two available variables: `Script`, which is the path to the script
    to run, `Vars`, which is the list of `environment_vars`, if configured.

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

Whether you use the `inline` option, or pass it a direct `script` or `scripts`,
it is important to understand a few things about how the shell-local
post-processor works to run it safely and easily. This understanding will save
you much time in the process.

### Once Per Builder

The `shell-local` script(s) you pass are run once per builder.  That means that
if you have an `amazon-ebs` builder and a `docker` builder, your script will be
run twice. If you have 3 builders, it will run 3 times, once for each builder.

### Interacting with Build Artifacts

In order to interact with build artifacts, you may want to use the [manifest
post-processor](/docs/post-processors/manifest.html). This will write the list
of files produced by a `builder` to a json file after each `builder` is run.

For example, if you wanted to package a file from the file builder into
a tarball, you might wright this:

```json
{
    "builders": [
        {
            "content": "Lorem ipsum dolor sit amet",
            "target": "dummy_artifact",
            "type": "file"
        }
    ],
    "post-processors": [
        [
            {
                "output": "manifest.json",
                "strip_path": true,
                "type": "manifest"
            },
            {
                "inline": [
                    "jq ".builds[].files[].name" manifest.json | xargs tar cfz artifacts.tgz"
                ],
                "type": "shell-local"
            }
        ]
    ]
}
```

This uses the [jq](https://stedolan.github.io/jq/) tool to extract all of the
file names from the manifest file and passes them to tar.

### Always Exit Intentionally

If any post-processor fails, the `packer build` stops and all interim artifacts
are cleaned up.

For a shell script, that means the script **must** exit with a zero code. You
*must* be extra careful to `exit 0` when necessary.

