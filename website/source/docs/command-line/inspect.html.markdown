---
description: |
    The `packer inspect` Packer command takes a template and outputs the various
    components a template defines. This can help you quickly learn about a template
    without having to dive into the JSON itself. The command will tell you things
    like what variables a template accepts, the builders it defines, the
    provisioners it defines and the order they'll run, and more.
layout: docs
page_title: 'Inspect - Command-Line'
...

# Command-Line: Inspect

The `packer inspect` Packer command takes a template and outputs the various
components a template defines. This can help you quickly learn about a template
without having to dive into the JSON itself. The command will tell you things
like what variables a template accepts, the builders it defines, the
provisioners it defines and the order they'll run, and more.

This command is extra useful when used with [machine-readable
output](/docs/command-line/machine-readable.html) enabled. The command outputs
the components in a way that is parseable by machines.

The command doesn't validate the actual configuration of the various components
(that is what the `validate` command is for), but it will validate the syntax of
your template by necessity.

## Usage Example

Given a basic template, here is an example of what the output might look like:

``` {.text}
$ packer inspect template.json
Variables and their defaults:

  aws_access_key =
  aws_secret_key =

Builders:

  amazon-ebs
  amazon-instance
  virtualbox-iso

Provisioners:

  shell
```
