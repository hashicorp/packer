---
layout: "docs"
page_title: "Shell Local Provisioner"
description: |-
  The shell-local Packer provisioner allows shell scripts to be run on machine
  running packer
---

# Shell Local Provisioner

Type: `shell-local`

The shell-local Packer provisioner allows shell scripts to be run on machine
running packer.

The primary use case for the shell-local provisioner is to allow for running an
image verification process such as [serverspec](http://serverspec.org/) prior to
creating an image without the need to run this from the machine you are
building.

Generally it is used as the last provisioner for a giving machine.

## Basic Example

The example below is fully functional.

```javascript
{
  "type": "shell-local",
  "inline": ["echo foo"]
}
```

## Configuration Reference

The reference of available configuration options is listed below. The only
required element is either "inline" or "script". Every other option is optional.

Required parameter:

* `inline` (array of strings) - This is an array of commands to execute.
  The commands are concatenated by newlines and turned into a single file,
  so they are all executed within the same context. This allows you to
  change directories in one command and use something in the directory in
  the next and so on.

Optional parameters:

* `inline_shebang` (string) - The
  [shebang](http://en.wikipedia.org/wiki/Shebang_%28Unix%29) value to use when
  running commands specified by `inline`. By default, this is `/bin/sh`.
  If you're not using `inline`, then this configuration has no effect.
