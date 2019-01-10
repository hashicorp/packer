---
description: |
    Packer Pre-processors are the components that are used to setup a environment.
    Example usages of pre-processors are setting up infrastructure or downloading a
    iso image and uploading it to a cloud provider for use in a builder.
layout: docs
page_title: 'Custom Pre-Processors - Extending'
sidebar_current: 'docs-extending-custom-pre-processors'
---

# Custom Pre-Processors

Packer Pre-processors are the components that are used to setup a environment,
for example setting up infrastructure.

Prior to reading this page, it is assumed you have read the page on [plugin
development basics](/docs/extending/plugins.html).

Pre-processor plugins implement the `packer.PreProcessor` interface and are
served using the `plugin.ServePreProcessor` function.

~&gt; **Warning!** This is an advanced topic. If you're new to Packer, we
recommend getting a bit more comfortable before you dive into writing plugins.

## The Interface

The interface that must be implemented for a pre-processor is the
`packer.PreProcessor` interface. It is reproduced below for reference. The
actual interface in the source code contains some basic documentation as well
explaining what each method should do.

``` go
type PreProcessor interface {
  Configure(interface{}) error
  PreProcess(Ui) error
}
```

### The "Configure" Method

The `Configure` method for each pre-processor is called early in the build
process to configure the pre-processor. The configuration is passed in as a raw
`interface{}`. The configure method is responsible for translating this
configuration into an internal structure, validating it, and returning any
errors.

For decoding the `interface{}` into a meaningful structure, the
[mapstructure](https://github.com/mitchellh/mapstructure) library is
recommended. Mapstructure will take an `interface{}` and decode it into an
arbitrarily complex struct. If there are any errors, it generates very
human-friendly errors that can be returned directly from the configure method.

While it is not actively enforced, **no side effects** should occur from
running the `Configure` method. Specifically, don't create files, don't create
network connections, etc. Configure's purpose is solely to setup internal state
and validate the configuration as much as possible.

`Configure` being run is not an indication that `PreProcess` will ever run. For
example, `packer validate` will run `Configure` to verify the configuration
validates, but will never actually run the build.

### The "PreProcess" Method

The `PreProcess` method is where the real work goes.

The result signature of this method is `error`. Each return value is explained
below:

-   `error` - Non-nil if there was an error in any way. If this is the case,
    the other two return values are ignored.
