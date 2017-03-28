---
layout: docs
sidebar_current: docs-commands-validate
page_title: packer validate - Commands
description: |-
  The `packer validate` Packer command is used to validate the syntax and
  configuration of a template. The command will return a zero exit status on
  success, and a non-zero exit status on failure. Additionally, if a template
  doesn't validate, any error messages will be outputted.
---

# `validate` Command

The `packer validate` Packer command is used to validate the syntax and
configuration of a [template](/docs/templates/index.html). The command
will return a zero exit status on success, and a non-zero exit status on
failure. Additionally, if a template doesn't validate, any error messages will
be outputted.

Example usage:

```text
$ packer validate my-template.json
Template validation failed. Errors are shown below.

Errors validating build 'vmware'. 1 error(s) occurred:

* Either a path or inline script must be specified.
```

## Options

- `-syntax-only` - Only the syntax of the template is checked. The configuration
  is not validated.
