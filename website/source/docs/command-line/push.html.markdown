---
layout: "docs"
page_title: "Push - Command-Line"
description: |-
  The `packer push` Packer command takes a template and pushes it to a build service that will automatically build this Packer template.
---

# Command-Line: Push

The `packer push` Packer command takes a template and pushes it to a Packer
build service such as [HashiCorp's Atlas](https://atlas.hashicorp.com). The
build service will automatically build your Packer template and expose the
artifacts.

External build services such as HashiCorp's Atlas make it easy to iterate on
Packer templates, especially when the builder you are running may not be easily
accessable (such as developing `qemu` builders on Mac or Windows).

For the `push` command to work, the [push configuration](/docs/templates/push.html)
must be completed within the template.

## Options

* `-token` - An access token for authenticating the push to the Packer build
  service such as Atlas. This can also be specified within the push
  configuration in the template.

## Examples

Push a Packer template:

```shell
$ packer push template.json
```

Push a Packer template with a custom token:

```shell
$ packer push -token ABCD1234 template.json
```
