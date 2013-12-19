---
layout: "docs"
page_title: "Vagrant Post-Processor"
---

# Vagrant Post-Processor

Type: `vagrant`

The Vagrant post-processor takes a build and converts the artifact
into a valid [Vagrant](http://www.vagrantup.com) box, if it can.
This lets you use Packer to automatically create arbitrarily complex
Vagrant boxes, and is in fact how the official boxes distributed by
Vagrant are created.

If you've never used a post-processor before, please read the
documentation on [using post-processors](/docs/templates/post-processors.html)
in templates. This knowledge will be expected for the remainder of
this document.

Because Vagrant boxes are [provider-specific](http://docs.vagrantup.com/v2/boxes/format.html),
the Vagrant post-processor is hardcoded to understand how to convert
the artifacts of certain builders into proper boxes for their
respective providers.

Currently, the Vagrant post-processor can create boxes for the following
providers.

* AWS
* DigitalOcean
* VirtualBox
* VMware

<div class="alert alert-block alert-info">
<strong>Support for additional providers</strong> is planned. If the
Vagrant post-processor doesn't support creating boxes for a provider you
care about, please help by contributing to Packer and adding support for it.
</div>

## Configuration

The simplest way to use the post-processor is to just enable it. No
configuration is required by default. This will mostly do what you expect
and will build functioning boxes for many of the built-in builders of
Packer.

However, if you want to configure things a bit more, the post-processor
does expose some configuration options. The available options are listed
below, with more details about certain options in following sections.

* `compression_level` (integer) - An integer repesenting the
  compression level to use when creating the Vagrant box.  Valid
  values range from 0 to 9, with 0 being no compression and 9 being
  the best compression. By default, compression is enabled at level 1.

* `include` (array of strings) - Paths to files to include in the
  Vagrant box. These files will each be copied into the top level directory
  of the Vagrant box (regardless of their paths). They can then be used
  from the Vagrantfile.

* `output` (string) - The full path to the box file that will be created
  by this post-processor. This is a
  [configuration template](/docs/templates/configuration-templates.html).
  The variable `Provider` is replaced by the Vagrant provider the box is for.
  The variable `ArtifactId` is replaced by the ID of the input artifact.
  The variable `BuildName` is replaced with the name of the build.
  By default, the value of this config is `packer_{{.BuildName}}_{{.Provider}}.box`.

* `vagrantfile_template` (string) - Path to a template to use for the
  Vagrantfile that is packaged with the box.
