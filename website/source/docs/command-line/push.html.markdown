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

!> The Packer build service will receive the raw copy of your Packer template
when you push. **If you have sensitive data in your Packer template, you should
move that data into Packer variables or environment variables!**

For the `push` command to work, the [push configuration](/docs/templates/push.html)
must be completed within the template.

## Options

* `-message` - A message to identify the purpose or changes in this Packer
  template much like a VCS commit message. This message will be passed to the
  Packer build service. This option is also available as a short option `-m`.

* `-token` - An access token for authenticating the push to the Packer build
  service such as Atlas. This can also be specified within the push
  configuration in the template.

* `-name` - The name of the build in the service. This typically
  looks like `hashicorp/precise64`.

## Examples

Push a Packer template:

```shell
$ packer push -m "Updating the apache version" template.json
```

Push a Packer template with a custom token:

```shell
$ packer push -token ABCD1234 template.json
```
