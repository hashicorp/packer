---
layout: "docs"
page_title: "Atlas Post-Processor"
description: |-
  The Atlas post-processor for Packer receives an artifact from a Packer build and uploads it to Atlas. Atlas hosts and serves artifacts, allowing you to version and distribute them in a simple way.
---

# Atlas Post-Processor

Type: `atlas`

The Atlas post-processor for Packer receives an artifact from a Packer build and uploads it to Atlas. [Atlas](https://atlas.hashicorp.com) hosts and serves artifacts, allowing you to version and distribute them in a simple way. 

## Workflow

To take full advantage of Packer and Atlas, it's important to understand the 
workflow for creating artifacts with Packer and storing them in Atlas using this post-processor. The goal of the Atlas post-processor is to streamline the distribution of public or private artifacts by hosting them in a central location in Atlas.

Here is an example workflow:

1. Packer builds an AMI with the [Amazon AMI builder](/docs/builders/amazon.html)
2. The `atlas` post-processor takes the resulting AMI and uploads it to Atlas. The `atlas` post-processor is configured with the name of the AMI, for example `hashicorp/foobar`, to create the artifact in Atlas or update the version if the artifact already exists
3. The new version is ready and available to be used in deployments with a tool like [Terraform](https://terraform.io)


## Configuration

The configuration allows you to specify and access the artifact in Atlas. 

### Required:

* `token` (string) - Your access token for the Atlas API.
  This can be generated on your [tokens page](https://atlas.hashicorp.com/settings/tokens). Alternatively you can export your Atlas token as an environmental variable and remove it from the configuration. 

* `artifact` (string) - The shorthand tag for your artifact that maps to
  Atlas, i.e `hashicorp/foobar` for `atlas.hashicorp.com/hashicorp/foobar`. You must 
  have access to the organization, hashicorp in this example, in order to add an artifact to 
  the organization in Atlas. 

* `artifact_type` (string) - For uploading AMIs to Atlas, `artifact_type` will always be `aws.ami`.
  This field must be defined because Atlas can host other artifact types, such as Vagrant boxes.

-> **Note:** If you want to upload Vagrant boxes to Atlas, for now use the [Vagrant Cloud post-processor](/docs/post-processors/vagrant-cloud.html).

### Optional:

* `atlas_url` (string) - Override the base URL for Atlas. This
is useful if you're using Atlas Enterprise in your own network. Defaults
to `https://atlas.hashicorp.com/api/v1`.

* `metadata` (map) - Send metadata about the artifact.

### Example Configuration

```javascript
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
        "source_ami": "ami-de0d9eb7",
        "instance_type": "t1.micro",
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
        "artifact_type": "aws.ami",
        "metadata": {
          "created_at": "{{timestamp}}"
        }
      }
    ]
}
```
