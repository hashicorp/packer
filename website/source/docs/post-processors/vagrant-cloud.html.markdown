---
layout: "docs"
page_title: "Vagrant Cloud Post-Processor"
description: |-
  The Packer Vagrant Cloud post-processor receives a Vagrant box from the `vagrant` post-processor and pushes it to Vagrant Cloud. Vagrant Cloud hosts and serves boxes to Vagrant, allowing you to version and distribute boxes to an organization in a simple way.
---

# Vagrant Cloud Post-Processor

~> Vagrant Cloud has been superseded by Atlas. Please use the [Atlas post-processor](/docs/post-processors/atlas.html) instead. Learn more about [Atlas](https://atlas.hashicorp.com/).

Type: `vagrant-cloud`

The Packer Vagrant Cloud post-processor receives a Vagrant box from the `vagrant`
post-processor and pushes it to Vagrant Cloud. [Vagrant Cloud](https://vagrantcloud.com)
hosts and serves boxes to Vagrant, allowing you to version and distribute
boxes to an organization in a simple way.

You'll need to be familiar with Vagrant Cloud, have an upgraded account
to enable box hosting, and be distributing your box via the [shorthand name](http://docs.vagrantup.com/v2/cli/box.html)
configuration.

## Workflow

It's important to understand the workflow that using this post-processor
enforces in order to take full advantage of Vagrant and Vagrant Cloud.

The use of this processor assume that you currently distribute, or plan
to distribute, boxes via Vagrant Cloud. It also assumes you create Vagrant
Boxes and deliver them to your team in some fashion.

Here is an example workflow:

1. You use Packer to build a Vagrant Box for the `virtualbox` provider
2. The `vagrant-cloud` post-processor is configured to point to the box `hashicorp/foobar` on Vagrant Cloud
via the `box_tag` configuration
2. The post-processor receives the box from the `vagrant` post-processor
3. It then creates the configured version, or verifies the existence of it, on Vagrant Cloud
4. A provider matching the name of the Vagrant provider is then created
5. The box is uploaded to Vagrant Cloud
6. The upload is verified
7. The version is released and available to users of the box


## Configuration

The configuration allows you to specify the target box that you have
access to on Vagrant Cloud, as well as authentication and version information.

### Required:

* `access_token` (string) - Your access token for the Vagrant Cloud API.
  This can be generated on your [tokens page](https://vagrantcloud.com/account/tokens).

* `box_tag` (string) - The shorthand tag for your box that maps to
   Vagrant Cloud, i.e `hashicorp/precise64` for `vagrantcloud.com/hashicorp/precise64`

* `version` (string) - The version number, typically incrementing a previous version.
  The version string is validated based on [Semantic Versioning](http://semver.org/). The string must match
  a pattern that could be semver, and doesn't validate that the version comes after
  your previous versions.


### Optional:

* `no_release` (string) - If set to true, does not release the version
on Vagrant Cloud, making it active. You can manually release the version
via the API or Web UI. Defaults to false.

* `vagrant_cloud_url` (string) - Override the base URL for Vagrant Cloud. This
is useful if you're using Vagrant Private Cloud in your own network. Defaults
to `https://vagrantcloud.com/api/v1`

* `version_description` (string) - Optionally markdown text used as a full-length
  and in-depth description of the version, typically for denoting changes introduced

* `box_download_url` (string) - Optional URL for a self-hosted box. If this is set
the box will not be uploaded to the Vagrant Cloud.

## Use with Vagrant Post-Processor

You'll need to use the Vagrant post-processor before using this post-processor.
An example configuration is below. Note the use of a doubly-nested array, which
ensures that the Vagrant Cloud post-processor is run after the Vagrant
post-processor.

```javascript
{
  "variables": {
    "version": "",
    "cloud_token": ""
  },
  "builders": [{
      // ...
  }],
  "post-processors": [
    [
      {
        "type": "vagrant",
        "include": ["image.iso"],
        "vagrantfile_template": "vagrantfile.tpl",
        "output": "proxycore_{{.Provider}}.box"
      },
      {
        "type": "vagrant-cloud",
        "box_tag": "hashicorp/precise64",
        "access_token": "{{user `cloud_token`}}",
        "version": "{{user `version`}}"
      }
    ]
  ]
}
```
