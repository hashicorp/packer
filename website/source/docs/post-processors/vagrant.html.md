---
description: |
    The Packer Vagrant post-processor takes a build and converts the artifact into
    a valid Vagrant box, if it can. This lets you use Packer to automatically
    create arbitrarily complex Vagrant boxes, and is in fact how the official boxes
    distributed by Vagrant are created.
layout: docs
page_title: 'Vagrant - Post-Processors'
sidebar_current: 'docs-post-processors-vagrant-box'
---

# Vagrant Post-Processor

Type: `vagrant`

The Packer Vagrant post-processor takes a build and converts the artifact into
a valid [Vagrant](https://www.vagrantup.com) box, if it can. This lets you use
Packer to automatically create arbitrarily complex Vagrant boxes, and is in
fact how the official boxes distributed by Vagrant are created.

If you've never used a post-processor before, please read the documentation on
[using post-processors](/docs/templates/post-processors.html) in templates.
This knowledge will be expected for the remainder of this document.

Because Vagrant boxes are
[provider-specific](https://docs.vagrantup.com/v2/boxes/format.html), the
Vagrant post-processor is hardcoded to understand how to convert the artifacts
of certain builders into proper boxes for their respective providers.

Currently, the Vagrant post-processor can create boxes for the following
providers.

-   AWS
-   Azure
-   DigitalOcean
-   Docker
-   Hyper-V
-   LXC
-   Parallels
-   QEMU
-   VirtualBox
-   VMware

-&gt; **Support for additional providers** is planned. If the Vagrant
post-processor doesn't support creating boxes for a provider you care about,
please help by contributing to Packer and adding support for it.

Please note that if you are using the Vagrant builder, then the Vagrant
post-processor is unnecesary because the output of the Vagrant builder is
already a Vagrant box; using this post-processor with the Vagrant builder will
cause your build to fail.

## Configuration

The simplest way to use the post-processor is to just enable it. No
configuration is required by default. This will mostly do what you expect and
will build functioning boxes for many of the built-in builders of Packer.

However, if you want to configure things a bit more, the post-processor does
expose some configuration options. The available options are listed below, with
more details about certain options in following sections.

-   `compression_level` (number) - An integer representing the compression
    level to use when creating the Vagrant box. Valid values range from 0 to 9,
    with 0 being no compression and 9 being the best compression. By default,
    compression is enabled at level 6.

-   `include` (array of strings) - Paths to files to include in the Vagrant
    box. These files will each be copied into the top level directory of the
    Vagrant box (regardless of their paths). They can then be used from the
    Vagrantfile.

-   `keep_input_artifact` (boolean) - When true, preserve the artifact we use to
    create the vagrant box. Defaults to `false`, except when you set a cloud
    provider (e.g. aws, azure, google, digitalocean). In these cases deleting
    the input artifact would render the vagrant box useless, so we always keep
    these artifacts -- even if you specifically set
    `"keep_input_artifact":false`

-   `output` (string) - The full path to the box file that will be created by
    this post-processor. This is a
    [template engine](/docs/templates/engine.html). Therefore, you may use user
    variables and template functions in this field. The following extra
    variables are also avilable in this engine:
     * `Provider`: The Vagrant provider the box is for
     * `ArtifactId`: The ID of the input artifact.
     * `BuildName`: The name of the build.

    By default, the value of this config is
    `packer_{{.BuildName}}_{{.Provider}}.box`.

-   `vagrantfile_template` (string) - Path to a template to use for the
    Vagrantfile that is packaged with the box.

-   `vagrantfile_template_generated` (boolean) - By default, Packer will
    exit with an error if the file specified using the
    `vagrantfile_template` variable is not found. However, under certain
    circumstances, it may be desirable to dynamically generate the
    Vagrantfile during the course of the build. Setting this variable to
    `true` skips the start up check and allows the user to script the
    creation of the Vagrantfile at some previous point in the build.
    Defaults to `false`.

## Provider-Specific Overrides

If you have a Packer template with multiple builder types within it, you may
want to configure the box creation for each type a little differently. For
example, the contents of the Vagrantfile for a Vagrant box for AWS might be
different from the contents of the Vagrantfile you want for VMware. The
post-processor lets you do this.

Specify overrides within the `override` configuration by provider name:

``` json
{
  "type": "vagrant",
  "compression_level": 1,
  "override": {
    "vmware": {
      "compression_level": 0
    }
  }
}
```

In the example above, the compression level will be set to 1 except for VMware,
where it will be set to 0.

The available provider names are:

-   `aws`
-   `azure`
-   `digitalocean`
-   `google`
-   `hyperv`
-   `parallels`
-   `libvirt`
-   `lxc`
-   `scaleway`
-   `virtualbox`
-   `vmware`
-   `docker`

## Input Artifacts

By default, Packer will delete the original input artifact, assuming you only
want the final Vagrant box as the result. If you wish to keep the input
artifact (the raw virtual machine, for example), then you must configure Packer
to keep it.

Please see the [documentation on input
artifacts](/docs/templates/post-processors.html#toc_2) for more information.

### Docker

Using a Docker input artifact will include a reference to the image in the
`Vagrantfile`. If the image tag is not specified in the post-processor, the
sha256 hash will be used.

The following Docker input artifacts are supported:

-   `docker` builder with `commit: true`, always uses the sha256 hash
-   `docker-import`
-   `docker-tag`
-   `docker-push`

### QEMU/libvirt

The `libvirt` provider supports QEMU artifacts built using any these
accelerators: none, kvm, tcg, or hvf.

### VMWare

If you are using the Vagrant post-processor with the `vmware-esxi` builder, you
must export the builder artifact locally; the Vagrant post-processor will
not work on remote artifacts.
