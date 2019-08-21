---
description: |
    The Vagrant Cloud post-processor enables the upload of Vagrant boxes to
    Vagrant Cloud.
layout: docs
page_title: 'Vagrant Cloud - Post-Processors'
sidebar_current: 'docs-post-processors-vagrant-cloud'
---

# Vagrant Cloud Post-Processor

Type: `vagrant-cloud`

[Vagrant Cloud](https://app.vagrantup.com/boxes/search) hosts and serves boxes
to Vagrant, allowing you to version and distribute boxes to an organization in a
simple way.

The Vagrant Cloud post-processor enables the upload of Vagrant boxes to Vagrant
Cloud. Currently, the Vagrant Cloud post-processor will accept and upload boxes
supplied to it from the [Vagrant](/docs/post-processors/vagrant.html) or
[Artifice](/docs/post-processors/artifice.html) post-processors and the
[Vagrant](/docs/builders/vagrant.html) builder.

You'll need to be familiar with Vagrant Cloud, have an upgraded account to
enable box hosting, and be distributing your box via the [shorthand
name](https://docs.vagrantup.com/v2/cli/box.html) configuration.

## Workflow

It's important to understand the workflow that using this post-processor
enforces in order to take full advantage of Vagrant and Vagrant Cloud.

The use of this processor assume that you currently distribute, or plan to
distribute, boxes via Vagrant Cloud. It also assumes you create Vagrant Boxes
and deliver them to your team in some fashion.

Here is an example workflow:

1.  You use Packer to build a Vagrant Box for the `virtualbox` provider
2.  The `vagrant-cloud` post-processor is configured to point to the box
    `hashicorp/foobar` on Vagrant Cloud via the `box_tag` configuration
3.  The post-processor receives the box from the `vagrant` post-processor
4.  It then creates the configured version, or verifies the existence of it, on
    Vagrant Cloud
5.  A provider matching the name of the Vagrant provider is then created
6.  The box is uploaded to Vagrant Cloud
7.  The upload is verified
8.  The version is released and available to users of the box

## Configuration

The configuration allows you to specify the target box that you have access to
on Vagrant Cloud, as well as authentication and version information.

### Required:

-   `access_token` (string) - Your access token for the Vagrant Cloud API. This
    can be generated on your [tokens
    page](https://app.vagrantup.com/settings/security). If not specified, the
    environment will be searched. First, `VAGRANT_CLOUD_TOKEN` is checked, and
    if nothing is found, finally `ATLAS_TOKEN` will be used.

-   `box_tag` (string) - The shorthand tag for your box that maps to Vagrant
    Cloud, for example `hashicorp/precise64`, which is short for
    `vagrantcloud.com/hashicorp/precise64`.

-   `version` (string) - The version number, typically incrementing a previous
    version. The version string is validated based on [Semantic
    Versioning](http://semver.org/). The string must match a pattern that could
    be semver, and doesn't validate that the version comes after your previous
    versions.

### Optional:

-   `no_release` (string) - If set to true, does not release the version on
    Vagrant Cloud, making it active. You can manually release the version via
    the API or Web UI. Defaults to false.

-   `vagrant_cloud_url` (string) - Override the base URL for Vagrant Cloud.
    This is useful if you're using Vagrant Private Cloud in your own network.
    Defaults to `https://vagrantcloud.com/api/v1`

-   `insecure_skip_tls_verify` (boolean) - If set to true *and* `vagrant_cloud_url`
    is set to something different than its default, it will set TLS InsecureSkipVerify
    to true. In other words, this will disable security checks of SSL. You may need
    to set this option to true if your host at vagrant_cloud_url is using a
    self-signed certificate.

-   `keep_input_artifact` (boolean) - When true, preserve the local box
    after uploading to Vagrant cloud. Defaults to `true`.

-   `version_description` (string) - Optionally markdown text used as a
    full-length and in-depth description of the version, typically for denoting
    changes introduced

-   `box_download_url` (string) - Optional URL for a self-hosted box. If this
    is set the box will not be uploaded to the Vagrant Cloud.

## Use with the Vagrant Post-Processor

An example configuration is shown below. Note the use of the nested array that
wraps both the Vagrant and Vagrant Cloud post-processors within the
post-processor section. Chaining the post-processors together in this way tells
Packer that the artifact produced by the Vagrant post-processor should be passed
directly to the Vagrant Cloud Post-Processor. It also sets the order in which
the post-processors should run.

Failure to chain the post-processors together in this way will result in the
wrong artifact being supplied to the Vagrant Cloud post-processor. This will
likely cause the Vagrant Cloud post-processor to error and fail.

``` json
{
  "variables": {
    "cloud_token": "{{ env `VAGRANT_CLOUD_TOKEN` }}",
    "version": "1.0.{{timestamp}}"
  },
  "post-processors": [
    {
      "type": "shell-local",
      "inline": ["echo Doing stuff..."]
    },
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

## Use with the Artifice Post-Processor

An example configuration is shown below. Note the use of the nested array that
wraps both the Artifice and Vagrant Cloud post-processors within the
post-processor section. Chaining the post-processors together in this way tells
Packer that the artifact produced by the Artifice post-processor should be
passed directly to the Vagrant Cloud Post-Processor. It also sets the order in
which the post-processors should run.

Failure to chain the post-processors together in this way will result in the
wrong artifact being supplied to the Vagrant Cloud post-processor. This will
likely cause the Vagrant Cloud post-processor to error and fail.

Note that the Vagrant box specified in the Artifice post-processor `files` array
must end in the `.box` extension. It must also be the first file in the array.
Additional files bundled by the Artifice post-processor will be ignored.

```json
{
  "variables": {
    "cloud_token": "{{ env `VAGRANT_CLOUD_TOKEN` }}",
  },

  "builders": [
    {
      "type": "null",
      "communicator": "none"
    }
  ],

  "post-processors": [
    {
      "type": "shell-local",
      "inline": ["echo Doing stuff..."]
    },
    [
      {
        "type": "artifice",
        "files": [
          "./path/to/my.box"
        ]
      },
      {
        "type": "vagrant-cloud",
        "box_tag": "myorganisation/mybox",
        "access_token": "{{user `cloud_token`}}",
        "version": "0.1.0",
      }
    ]
  ]
}
```
