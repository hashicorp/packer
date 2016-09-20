---
description: |
    Within the template, the builders section contains an array of all the builders
    that Packer should use to generate machine images for the template.
layout: docs
page_title: 'Templates: Builders'
...

# Templates: Builders

Within the template, the builders section contains an array of all the builders
that Packer should use to generate a machine images for the template.

Builders are responsible for creating machines and generating images from them
for various platforms. For example, there are separate builders for EC2, VMware,
VirtualBox, etc. Packer comes with many builders by default, and can also be
extended to add new builders.

This documentation page will cover how to configure a builder in a template. The
specific configuration options available for each builder, however, must be
referenced from the documentation for that specific builder.

Within a template, a section of builder definitions looks like this:

``` {.javascript}
{
  "builders": [
    // ... one or more builder definitions here
  ]
}
```

## Builder Definition

A single builder definition maps to exactly one
[build](/docs/basics/terminology.html#term-build). A builder definition is a
JSON object that requires at least a `type` key. The `type` is the name of the
builder that will be used to create a machine image for the build.

In addition to the `type`, other keys configure the builder itself. For example,
the AWS builder requires an `access_key`, `secret_key`, and some other settings.
These are placed directly within the builder definition.

An example builder definition is shown below, in this case configuring the AWS
builder:

``` {.javascript}
{
  "type": "amazon-ebs",
  "access_key": "...",
  "secret_key": "..."
}
```

## Named Builds

Each build in Packer has a name. By default, the name is just the name of the
builder being used. In general, this is good enough. Names only serve as an
indicator in the output of what is happening. If you want, however, you can
specify a custom name using the `name` key within the builder definition.

This is particularly useful if you have multiple builds defined that use the
same underlying builder. In this case, you must specify a name for at least one
of them since the names must be unique.

## Communicators

Every build is associated with a single
[communicator](/docs/templates/communicator.html). Communicators are used to
establish a connection for provisioning a remote machine (such as an AWS
instance or local virtual machine).

All the examples for the various builders show some communicator (usually SSH),
but the communicators are highly customizable so we recommend reading the
[communicator documentation](/docs/templates/communicator.html).
