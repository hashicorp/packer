---
description: |
    Up to this point in the guide, you have been running Packer on your local
    machine to build and provision images on AWS and DigitalOcean. However, you can
    use Atlas by HashiCorp to both run Packer builds remotely and store the output
    of builds.
layout: intro
next_title: Next Steps
next_url: '/intro/getting-started/next.html'
page_title: Remote Builds and Storage
prev_url: '/intro/getting-started/vagrant.html'
...

# Remote Builds and Storage

Up to this point in the guide, you have been running Packer on your local
machine to build and provision images on AWS and DigitalOcean. However, you can
use [Atlas by HashiCorp](https://atlas.hashicorp.com) to run Packer builds
remotely and store the output of builds.

## Why Build Remotely?

By building remotely, you can move access credentials off of developer machines,
release local machines from long-running Packer processes, and automatically
start Packer builds from trigger sources such as `vagrant push`, a version
control system, or CI tool.

## Run Packer Builds Remotely

To run Packer remotely, there are two changes that must be made to the Packer
template. The first is the addition of the `push`
[configuration](https://www.packer.io/docs/templates/push.html), which sends the
Packer template to Atlas so it can run Packer remotely. The second modification
is updating the variables section to read variables from the Atlas environment
rather than the local environment. Remove the `post-processors` section for now
if it is still in your template.

``` {.javascript}
{
  "variables": {
    "aws_access_key": "{{env `aws_access_key`}}",
    "aws_secret_key": "{{env `aws_secret_key`}}"
  },
  "builders": [{
    "type": "amazon-ebs",
    "access_key": "{{user `aws_access_key`}}",
    "secret_key": "{{user `aws_secret_key`}}",
    "region": "us-east-1",
    "source_ami": "ami-9eaa1cf6",
    "instance_type": "t2.micro",
    "ssh_username": "ubuntu",
    "ami_name": "packer-example {{timestamp}}"
  }],
  "provisioners": [{
    "type": "shell",
    "inline": [
      "sleep 30",
      "sudo apt-get update",
      "sudo apt-get install -y redis-server"
    ]
  }],
  "push": {
    "name": "ATLAS_USERNAME/packer-tutorial"
  }
}
```

To get an Atlas username, [create an account
here](https://atlas.hashicorp.com/account/new?utm_source=oss&utm_medium=getting-started&utm_campaign=packer).
Replace "ATLAS\_USERNAME" with your username, then run
`packer push -create example.json` to send the configuration to Atlas, which
automatically starts the build.

This build will fail since neither `aws_access_key` or `aws_secret_key` are set
in the Atlas environment. To set environment variables in Atlas, navigate to
the [Builds tab](https://atlas.hashicorp.com/builds), click the
"packer-tutorial" build configuration that was just created, and then click
'variables' in the left navigation. Set `aws_access_key` and `aws_secret_key`
with their respective values. Now restart the Packer build by either clicking
'rebuild' in the Atlas UI or by running `packer push example.json` again. Now
when you click on the active build, you can view the logs in real-time.

-&gt; **Note:** Whenever a change is made to the Packer template, you must
`packer push` to update the configuration in Atlas.

## Store Packer Outputs

Now we have Atlas building an AMI with Redis pre-configured. This is great, but
it's even better to store and version the AMI output so it can be easily
deployed by a tool like [Terraform](https://www.terraform.io). The `atlas`
[post-processor](/docs/post-processors/atlas.html) makes this process simple:

``` {.javascript}
{
  "variables": ["..."],
  "builders": ["..."],
  "provisioners": ["..."],
  "push": ["..."],
  "post-processors": [{
    "type": "atlas",
    "artifact": "ATLAS_USERNAME/packer-tutorial",
    "artifact_type": "amazon.image"
  }]
}
```

Update the `post-processors` block with your Atlas username, then
`packer push example.json` and watch the build kick off in Atlas! When the build
completes, the resulting artifact will be saved and stored in Atlas.
