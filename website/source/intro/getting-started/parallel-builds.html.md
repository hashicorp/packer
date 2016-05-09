---
description: |
    So far we've shown how Packer can automatically build an image and provision it.
    This on its own is already quite powerful. But Packer can do better than that.
    Packer can create multiple images for multiple platforms in parallel, all
    configured from a single template.
layout: intro
next_title: Vagrant Boxes
next_url: '/intro/getting-started/vagrant.html'
page_title: Parallel Builds
prev_url: '/intro/getting-started/provision.html'
...

# Parallel Builds

So far we've shown how Packer can automatically build an image and provision it.
This on its own is already quite powerful. But Packer can do better than that.
Packer can create multiple images for multiple platforms *in parallel*, all
configured from a single template.

This is a very useful and important feature of Packer. As an example, Packer is
able to make an AMI and a VMware virtual machine in parallel provisioned with
the *same scripts*, resulting in near-identical images. The AMI can be used for
production, the VMware machine can be used for development. Or, another example,
if you're using Packer to build [software
appliances](https://en.wikipedia.org/wiki/Software_appliance), then you can build
the appliance for every supported platform all in parallel, all configured from
a single template.

Once you start taking advantage of this feature, the possibilities begin to
unfold in front of you.

Continuing on the example in this getting started guide, we'll build a
[DigitalOcean](http://www.digitalocean.com) image as well as an AMI. Both will
be near-identical: bare bones Ubuntu OS with Redis pre-installed. However, since
we're building for both platforms, you have the option of whether you want to
use the AMI, or the DigitalOcean snapshot. Or use both.

## Setting Up DigitalOcean

[DigitalOcean](https://www.digitalocean.com/) is a relatively new, but very
popular VPS provider that has popped up. They have a quality offering of high
performance, low cost VPS servers. We'll be building a DigitalOcean snapshot for
this example.

In order to do this, you'll need an account with DigitalOcean. [Sign up for an
account now](https://www.digitalocean.com/). It is free to sign up. Because the
"droplets" (servers) are charged hourly, you *will* be charged $0.01 for every
image you create with Packer. If you're not okay with this, just follow along.

!&gt; **Warning!** You *will* be charged $0.01 by DigitalOcean per image
created with Packer because of the time the "droplet" is running.

Once you sign up for an account, grab your API token from the [DigitalOcean API
access page](https://cloud.digitalocean.com/settings/applications). Save these
values somewhere; you'll need them in a second.

## Modifying the Template

We now have to modify the template to add DigitalOcean to it. Modify the
template we've been using and add the following JSON object to the `builders`
array.

``` {.javascript}
{
  "type": "digitalocean",
  "api_token": "{{user `do_api_token`}}",
  "image": "ubuntu-14-04-x64",
  "region": "nyc3",
  "size": "512mb"
}
```

You'll also need to modify the `variables` section of the template to include
the access keys for DigitalOcean.

``` {.javascript}
"variables": {
  "do_api_token": "",
  // ...
}
```

The entire template should now look like this:

``` {.javascript}
{
  "variables": {
    "aws_access_key": "",
    "aws_secret_key": "",
    "do_api_token": ""
  },
  "builders": [{
    "type": "amazon-ebs",
    "access_key": "{{user `aws_access_key`}}",
    "secret_key": "{{user `aws_secret_key`}}",
    "region": "us-east-1",
    "source_ami": "ami-fce3c696",
    "instance_type": "t2.micro",
    "ssh_username": "ubuntu",
    "ami_name": "packer-example {{timestamp}}"
  },{
    "type": "digitalocean",
    "api_token": "{{user `do_api_token`}}",
    "image": "ubuntu-14-04-x64",
    "region": "nyc3",
    "size": "512mb"
  }],
  "provisioners": [{
    "type": "shell",
    "inline": [
      "sleep 30",
      "sudo apt-get update",
      "sudo apt-get install -y redis-server"
    ]
  }]
}
```

Additional builders are simply added to the `builders` array in the template.
This tells Packer to build multiple images. The builder `type` values don't even
need to be different! In fact, if you wanted to build multiple AMIs, you can do
that as long as you specify a unique `name` for each build.

Validate the template with `packer validate`. This is always a good practice.

-&gt; **Note:** If you're looking for more **DigitalOcean configuration
options**, you can find them on the [DigitalOcean Builder
page](/docs/builders/digitalocean.html) in the documentation. The documentation
is more of a reference manual that contains a listing of all the available
configuration options.

## Build

Now run `packer build` with your user variables. The output is too verbose to
include all of it, but a portion of it is reproduced below. Note that the
ordering and wording of the lines may be slightly different, but the effect is
the same.

``` {.text}
$ packer build \
    -var 'aws_access_key=YOUR ACCESS KEY' \
    -var 'aws_secret_key=YOUR SECRET KEY' \
    -var 'do_api_token=YOUR API TOKEN' \
    example.json
==> amazon-ebs: amazon-ebs output will be in this color.
==> digitalocean: digitalocean output will be in this color.

==> digitalocean: Creating temporary ssh key for droplet...
==> amazon-ebs: Creating temporary keypair for this instance...
==> amazon-ebs: Creating temporary security group for this instance...
==> digitalocean: Creating droplet...
==> amazon-ebs: Authorizing SSH access on the temporary security group...
==> amazon-ebs: Launching a source AWS instance...
==> digitalocean: Waiting for droplet to become active...
==> amazon-ebs: Waiting for instance to become ready...
==> digitalocean: Connecting to the droplet via SSH...
==> amazon-ebs: Connecting to the instance via SSH...
...
==> Builds finished. The artifacts of successful builds are:
--> amazon-ebs: AMIs were created:

us-east-1: ami-376d1d5e
--> digitalocean: A snapshot was created: packer-1371870364
```

As you can see, Packer builds both the Amazon and DigitalOcean images in
parallel. It outputs information about each in different colors (although you
can't see that in the block above) so that it is easy to identify.

At the end of the build, Packer outputs both of the artifacts created (an AMI
and a DigitalOcean snapshot). Both images created are bare bones Ubuntu
installations with Redis pre-installed.
