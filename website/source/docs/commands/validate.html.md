---
description: |
    The `packer validate` Packer command is used to validate the syntax and
    configuration of a template. The command will return a zero exit status on
    success, and a non-zero exit status on failure. Additionally, if a template
    doesn't validate, any error messages will be outputted.
layout: docs
page_title: 'packer validate - Commands'
sidebar_current: 'docs-commands-validate'
---

# `validate` Command

The `packer validate` Packer command is used to validate the syntax and
configuration of a [template](/docs/templates/index.html). The command will
return a zero exit status on success, and a non-zero exit status on failure.
Additionally, if a template doesn't validate, any error messages will be
outputted.

Example usage:

``` text
$ packer validate my-template.json
Template validation failed. Errors are shown below.

Errors validating build 'vmware'. 1 error(s) occurred:

* Either a path or inline script must be specified.
```

## Options

-   `-syntax-only` - Only the syntax of the template is checked. The
    configuration is not validated.

-   `-except=foo,bar,baz` - Builds all the builds except those with the given
    comma-separated names. Build names by default are the names of their
    builders, unless a specific `name` attribute is specified within the
    configuration.

-   `-only=foo,bar,baz` - Only build the builds with the given comma-separated
    names. Build names by default are the names of their builders, unless a
    specific `name` attribute is specified within the configuration.

-   `-var` - Set a variable in your packer template. This option can be used
    multiple times. This is useful for setting version numbers for your build.

-   `-var-file` - Set template variables from a file.
