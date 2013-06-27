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

Because Vagrant boxes are [provider-specific](#),
the Vagrant post-processor is hardcoded to understand how to convert
the artifacts of certain builders into proper boxes for their
respective providers.

Currently, the Vagrant post-processor can create boxes for the following
providers.

* AWS
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

* `output` (string) - The full path to the box file that will be created
  by this post-processor. This is a
  [configuration template](/docs/templates/configuration-templates.html).
  The variable `Provider` is replaced by the Vagrant provider the box is for.
  By default, the value of this config is `packer_{{.Provider}}.box`.

* `aws`, `virtualbox`, or `vmware` (objects) - These are used to configure
  the specific options for certain providers. A reference of available
  configuration parameters for each is in the section below.

### AWS Provider

The AWS provider itself can be configured with specific options:

* `vagrantfile_template` (string) - Path to a template to use for the
  Vagrantfile that is packaged with the box. The contents of the file must be a valid Go
  [text template](http://golang.org/pkg/text/template). By default
  this is a template that simply sets the AMIs for the various regions
  of the AWS build.

The `vagrantfile_template` has the `Images` variable which is a map
of region (string) to AMI ID (string). An example Vagrantfile template for
AWS is shown below. The example simply sets the AMI for each region.

```
Vagrant.configure("2") do |config|
  config.vm.provider "aws" do |aws|
    {{ range $region, $ami := .Images }}
	aws.region_config "{{ $region }}", ami: "{{ $ami }}"
	{{ end }}
  end
end
```

### VirtualBox Provider

The VirtualBox provider itself can be configured with specific options:

* `vagrantfile_template` (string) - Path to a template to use for the
  Vagrantfile that is packaged with the box. The contents of the file must be a valid Go
  [text template](http://golang.org/pkg/text/template). By default this is
  a template that just sets the base MAC address so that networking works.

The `vagrantfile_template` has the `BaseMACAddress` variable which is a string
containing the MAC address of the first network interface. This must be set
in the Vagrantfile for networking to work properly with Vagrant. An example
Vagrantfile template is shown below:

```
TODO
```

### VMware Provider

The VMware provider itself can be configured with specific options:

* `vagrantfile_template` (string) - Path to a template to use for the
  Vagrantfile that is packaged with the box. The contents of the file must be a valid Go
  [text template](http://golang.org/pkg/text/template). By default no
  Vagrantfile is packaged with the box. Note that currently no variables
  are available in the template, but this may change in the future.
