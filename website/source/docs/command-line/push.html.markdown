---
layout: "docs"
page_title: "Push - Command-Line"
description: |-
  The `packer push` Packer command takes a template and pushes it to a build service that will automatically build this Packer template.
---

# Command-Line: Push

The `packer push` Packer command takes a template and pushes it to a build
service. The build service will automatically build your Packer template and
expose the artifacts.

This command currently only sends templates to
[Atlas](https://atlas.hashicorp.com) by HashiCorp, but the command will
be pluggable in the future with alternate implementations.

External build services such as Atlas make it easy to iterate on Packer
templates, especially when the builder you're running may not be easily
accessable (such as developing `qemu` builders on Mac or Windows).

For the `push` command to work, the
[push configuration](/docs/templates/push.html)
must be completed within the template.

## Options

* `-create=true` - If the build configuration matching the name of the push
  doesn't exist, it will be created if this is true. This defaults to true.

* `-token=FOO` - An access token for authenticating the push. This can also
  be specified within the push configuration in the template. By setting this
  in the template, you can take advantage of user variables.
