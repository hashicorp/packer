---
description: |
    Packer also has the ability to take the results of a builder (such as an AMI or
    plain VMware image) and turn it into a Vagrant box.
layout: intro
next_title: Remote Builds and Storage
next_url: '/intro/getting-started/remote-builds.html'
page_title: Vagrant Boxes
prev_url: '/intro/getting-started/parallel-builds.html'
...

# Vagrant Boxes

Packer also has the ability to take the results of a builder (such as an AMI or
plain VMware image) and turn it into a [Vagrant](https://www.vagrantup.com) box.

This is done using [post-processors](/docs/templates/post-processors.html).
These take an artifact created by a previous builder or post-processor and
transforms it into a new one. In the case of the Vagrant post-processor, it
takes an artifact from a builder and transforms it into a Vagrant box file.

Post-processors are a generally very useful concept. While the example on this
getting-started page will be creating Vagrant images, post-processors have many
interesting use cases. For example, you can write a post-processor to compress
artifacts, upload them, test them, etc.

Let's modify our template to use the Vagrant post-processor to turn our AWS AMI
into a Vagrant box usable with the [vagrant-aws
plugin](https://github.com/mitchellh/vagrant-aws). If you followed along in the
previous page and setup DigitalOcean, Packer can't currently make Vagrant boxes
for DigitalOcean, but will be able to soon.

## Enabling the Post-Processor

Post-processors are added in the `post-processors` section of a template, which
we haven't created yet. Modify your `example.json` template and add the section.
Your template should look like the following:

``` {.javascript}
{
  "builders": ["..."],
  "provisioners": ["..."],
  "post-processors": ["vagrant"]
}
```

In this case, we're enabling a single post-processor named "vagrant". This
post-processor is built-in to Packer and will create Vagrant boxes. You can
always create [new post-processors](/docs/extend/post-processor.html), however.
The details on configuring post-processors is covered in the
[post-processors](/docs/templates/post-processors.html) documentation.

Validate the configuration using `packer validate`.

## Using the Post-Processor

Just run a normal `packer build` and it will now use the post-processor. Since
Packer can't currently make a Vagrant box for DigitalOcean anyways, I recommend
passing the `-only=amazon-ebs` flag to `packer build` so it only builds the AMI.
The command should look like the following:

``` {.text}
$ packer build -only=amazon-ebs example.json
```

As you watch the output, you'll notice at the end in the artifact listing that a
Vagrant box was made (by default at `packer_aws.box` in the current directory).
Success!

But where did the AMI go? When using post-processors, Vagrant removes
intermediary artifacts since they're usually not wanted. Only the final artifact
is preserved. This behavior can be changed, of course. Changing this behavior is
covered [in the documentation](/docs/templates/post-processors.html).

Typically when removing intermediary artifacts, the actual underlying files or
resources of the artifact are also removed. For example, when building a VMware
image, if you turn it into a Vagrant box, the files of the VMware image will be
deleted since they were compressed into the Vagrant box. With creating AWS
images, however, the AMI is kept around, since Vagrant needs it to function.
