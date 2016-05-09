---
description: |
    The Atlas post-processor for Packer receives an artifact from a Packer build and
    uploads it to Atlas. Atlas hosts and serves artifacts, allowing you to version
    and distribute them in a simple way.
layout: docs
page_title: 'Atlas Post-Processor'
...

# Atlas Post-Processor

Type: `atlas`

The Atlas post-processor uploads artifacts from your packer builds to Atlas for
hosting. Artifacts hosted in Atlas are automatically made available for use
with Vagrant and Terraform, and Atlas provides additional features for managing
versions and releases. [Learn more about packer in
Atlas.](https://atlas.hashicorp.com/help/packer/features)

You can also use the push command to [run packer builds in
Atlas](/docs/command-line/push.html). The push command and Atlas post-processor
can be used together or independently.

## Workflow

To take full advantage of Packer and Atlas, it's important to understand the
workflow for creating artifacts with Packer and storing them in Atlas using this
post-processor. The goal of the Atlas post-processor is to streamline the
distribution of public or private artifacts by hosting them in a central
location in Atlas.

Here is an example workflow:

1.  Packer builds an AMI with the [Amazon AMI
    builder](/docs/builders/amazon.html)
2.  The `atlas` post-processor takes the resulting AMI and uploads it to Atlas.
    The `atlas` post-processor is configured with the name of the AMI, for
    example `hashicorp/foobar`, to create the artifact in Atlas or update the
    version if the artifact already exists
3.  The new version is ready and available to be used in deployments with a
    tool like [Terraform](https://www.terraform.io)

## Configuration

The configuration allows you to specify and access the artifact in Atlas.

### Required:

-   `token` (string) - Your access token for the Atlas API.

-&gt; Login to Atlas to [generate an Atlas
Token](https://atlas.hashicorp.com/settings/tokens). The most convenient way to
configure your token is to set it to the `ATLAS_TOKEN` environment variable, but
you can also use `token` configuration option.

-   `artifact` (string) - The shorthand tag for your artifact that maps to
    Atlas, i.e `hashicorp/foobar` for `atlas.hashicorp.com/hashicorp/foobar`.
    You must have access to the organization—hashicorp in this example—in order
    to add an artifact to the organization in Atlas.

-   `artifact_type` (string) - For uploading artifacts to Atlas.
    `artifact_type` can be set to any unique identifier, however, the following
    are recommended for consistency - `amazon.image`, `digitalocean.image`,
    `docker.image`, `googlecompute.image`, `openstack.image`,
    `parallels.image`, `qemu.image`, `virtualbox.image`, `vmware.image`,
    `custom.image`, and `vagrant.box`.

### Optional:

-   `atlas_url` (string) - Override the base URL for Atlas. This is useful if
    you're using Atlas Enterprise in your own network. Defaults to
    `https://atlas.hashicorp.com/api/v1`.

-   `metadata` (map) - Send metadata about the artifact. If the artifact type
    is `vagrant.box`, you must specify a `provider` metadata about what
    provider to use.

    -   `description` (string) - Inside the metadata blob you can add a information
        about the uploaded artifact to Atlas. This will be reflected in the box
        description on Atlas.

    -   `provider` (string) - Used by Atlas to help determine, what should be used
        to run the artifact.

    -   `version` (string) - Used by Atlas to give a semantic version to the
        uploaded artifact.

## Environment Variables

-   `ATLAS_CAFILE` (path) - This should be a path to an X.509 PEM-encoded public key. If specified, this will be used to validate the certificate authority that signed certificates used by an Atlas installation.

-   `ATLAS_CAPATH` - This should be a path which contains an X.509 PEM-encoded public key file. If specified, this will be used to validate the certificate authority that signed certificates used by an Atlas installation.

### Example Configuration

``` {.javascript}
{
    "variables": {
        "aws_access_key": "ACCESS_KEY_HERE",
        "aws_secret_key": "SECRET_KEY_HERE",
        "atlas_token": "ATLAS_TOKEN_HERE"
    },
    "builders": [{
        "type": "amazon-ebs",
        "access_key": "{{user `aws_access_key`}}",
        "secret_key": "{{user `aws_secret_key`}}",
        "region": "us-east-1",
        "source_ami": "ami-fce3c696",
        "instance_type": "t2.micro",
        "ssh_username": "ubuntu",
        "ami_name": "atlas-example {{timestamp}}"
    }],
    "provisioners": [
    {
        "type": "shell",
        "inline": [
            "sleep 30",
            "sudo apt-get update",
            "sudo apt-get install apache2 -y"
        ]
    }],
    "post-processors": [
      {
        "type": "atlas",
        "token": "{{user `atlas_token`}}",
        "artifact": "hashicorp/foobar",
        "artifact_type": "amazon.image",
        "metadata": {
          "created_at": "{{timestamp}}"
        }
      }
    ]
}
```

More information on the correct configuration of the `amazon-ebs` builder in this example can be found in the [amazon-ebs builder documentation](/docs/builders/amazon-ebs.html).
