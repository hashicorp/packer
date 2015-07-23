---
description: |
    Packer Post-processors are the components of Packer that transform one artifact
    into another, for example by compressing files, or uploading them.
layout: docs
page_title: 'Custom Post-Processor Development'
...

# Custom Post-Processor Development

Packer Post-processors are the components of Packer that transform one artifact
into another, for example by compressing files, or uploading them.

In the compression example, the transformation would be taking an artifact with
a set of files, compressing those files, and returning a new artifact with only
a single file (the compressed archive). For the upload example, the
transformation would be taking an artifact with some set of files, uploading
those files, and returning an artifact with a single ID: the URL of the upload.

Prior to reading this page, it is assumed you have read the page on [plugin
development basics](/docs/extend/developing-plugins.html).

Post-processor plugins implement the `packer.PostProcessor` interface and are
served using the `plugin.ServePostProcessor` function.

\~&gt; **Warning!** This is an advanced topic. If you're new to Packer, we
recommend getting a bit more comfortable before you dive into writing plugins.

## The Interface

The interface that must be implemented for a post-processor is the
`packer.PostProcessor` interface. It is reproduced below for easy reference. The
actual interface in the source code contains some basic documentation as well
explaining what each method should do.

``` {.go}
type PostProcessor interface {
  Configure(interface{}) error
  PostProcess(Ui, Artifact) (a Artifact, keep bool, err error)
}
```

### The "Configure" Method

The `Configure` method for each post-processor is called early in the build
process to configure the post-processor. The configuration is passed in as a raw
`interface{}`. The configure method is responsible for translating this
configuration into an internal structure, validating it, and returning any
errors.

For decoding the `interface{}` into a meaningful structure, the
[mapstructure](https://github.com/mitchellh/mapstructure) library is
recommended. Mapstructure will take an `interface{}` and decode it into an
arbitrarily complex struct. If there are any errors, it generates very
human-friendly errors that can be returned directly from the configure method.

While it is not actively enforced, **no side effects** should occur from running
the `Configure` method. Specifically, don't create files, don't create network
connections, etc. Configure's purpose is solely to setup internal state and
validate the configuration as much as possible.

`Configure` being run is not an indication that `PostProcess` will ever run. For
example, `packer validate` will run `Configure` to verify the configuration
validates, but will never actually run the build.

### The "PostProcess" Method

The `PostProcess` method is where the real work goes. PostProcess is responsible
for taking one `packer.Artifact` implementation, and transforming it into
another.

When we say "transform," we don't mean actually modifying the existing
`packer.Artifact` value itself. We mean taking the contents of the artifact and
creating a new artifact from that. For example, if we were creating a "compress"
post-processor that is responsible for compressing files, the transformation
would be taking the `Files()` from the original artifact, compressing them, and
creating a new artifact with a single file: the compressed archive.

The result signature of this method is `(Artifact, bool, error)`. Each return
value is explained below:

-   `Artifact` - The newly created artifact if no errors occurred.
-   `bool` - If true, the input artifact will forcefully be kept. By default,
    Packer typically deletes all input artifacts, since the user doesn't
    generally want intermediary artifacts. However, some post-processors depend
    on the previous artifact existing. If this is `true`, it forces packer to
    keep the artifact around.
-   `error` - Non-nil if there was an error in any way. If this is the case, the
    other two return values are ignored.
