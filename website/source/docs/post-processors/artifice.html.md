---
description: |
    The artifice post-processor overrides the artifact list from an upstream builder
    or post-processor. All downstream post-processors will see the new artifacts you
    specify. The primary use-case is to build artifacts inside a packer builder --
    for example, spinning up an EC2 instance to build a docker container -- and then
    extracting the docker container and throwing away the EC2 instance.
layout: docs
page_title: 'Artifice Post-Processor'
...

# Artifice Post-Processor

Type: `artifice`

The artifice post-processor overrides the artifact list from an upstream builder
or post-processor. All downstream post-processors will see the new artifacts you
specify. The primary use-case is to build artifacts inside a packer builder --
for example, spinning up an EC2 instance to build a docker container -- and then
extracting the docker container and throwing away the EC2 instance.

After overriding the artifact with artifice, you can use it with other
post-processors like
[compress](https://www.packer.io/docs/post-processors/compress.html),
[docker-push](https://www.packer.io/docs/post-processors/docker-push.html),
[Atlas](https://www.packer.io/docs/post-processors/atlas.html), or a third-party
post-processor.

Artifice allows you to use the familiar packer workflow to create a fresh,
stateless build environment for each build on the infrastructure of your
choosing. You can use this to build just about anything: buildpacks, containers,
jars, binaries, tarballs, msi installers, and more.

## Workflow

Artifice helps you tie together a few other packer features:

-   A builder, which spins up a VM (or container) to build your artifact
-   A provisioner, which performs the steps to create your artifact
-   A file provisioner, which downloads the artifact from the VM
-   The artifice post-processor, which identifies which files have been
    downloaded from the VM
-   Additional post-processors, which push the artifact to Atlas, Docker
    hub, etc.

You will want to perform as much work as possible inside the VM. Ideally the
only other post-processor you need after artifice is one that uploads your
artifact to the appropriate repository.

## Configuration

The configuration allows you to specify which files comprise your artifact.

### Required:

-   `files` (array of strings) - A list of files that comprise your artifact.
    These files must exist on your local disk after the provisioning phase of
    packer is complete. These will replace any of the builder's original
    artifacts (such as a VM snapshot).

### Example Configuration

This minimal example:

1.  Spins up a cloned VMware virtual machine
2.  Installs a [consul](https://www.consul.io/) release
3.  Downloads the consul binary
4.  Packages it into a `.tar.gz` file
5.  Uploads it to Atlas.

VMX is a fast way to build and test locally, but you can easily substitute
another builder.

``` {.javascript}
{
  "builders": [
    {
      "type": "vmware-vmx",
      "source_path": "/opt/ubuntu-1404-vmware.vmx",
      "ssh_username": "vagrant",
      "ssh_password": "vagrant",
      "shutdown_command": "sudo shutdown -h now",
      "headless":"true",
      "skip_compaction":"true"
    }
  ],
  "provisioners": [
    {
      "type": "shell",
      "inline": [
        "sudo apt-get install -y python-pip",
        "sudo pip install ifs",
        "sudo ifs install consul --version=0.5.2"
      ]
    },
    {
      "type": "file",
      "source": "/usr/local/bin/consul",
      "destination": "consul",
      "direction": "download"
    }
  ],
  "post-processors": [
    [
      {
        "type": "artifice",
        "files": ["consul"]
      },
      {
        "type": "compress",
        "output": "consul-0.5.2.tar.gz"
      },
      {
        "type":"atlas",
        "artifact": "hashicorp/consul",
        "artifact_type": "archive"
      }
    ]
  ]
}
```

**Notice that there are two sets of square brackets in the post-processor
section.** This creates a post-processor chain, where the output of the
proceeding artifact is passed to subsequent post-processors. If you use only one
set of square braces the post-processors will run individually against the build
artifact (the vmx file in this case) and it will not have the desired result.

      "post-processors": [
        [       <--- Start post-processor chain
          {
            "type": "artifice",
            "files": ["consul"]
          },
          {
            "type": "atlas",
            ...
          }
        ],      <--- End post-processor chain
        {
          "type":"compress"  <-- Standalone post-processor
        }
      ]

You can create multiple post-processor chains to handle multiple builders (for
example, building linux and windows binaries during the same build).
