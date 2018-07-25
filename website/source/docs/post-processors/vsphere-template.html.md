---
description: |
    The Packer vSphere Template post-processor takes an artifact from the VMware-iso builder, built on ESXi (i.e. remote)
    or an artifact from the vSphere post-processor and allows to mark a VM as a template and leaving it in a path of choice.
layout: docs
page_title: 'vSphere Template - Post-Processors'
sidebar_current: 'docs-post-processors-vSphere-template'
---

# vSphere Template Post-Processor

Type: `vsphere-template`

The Packer vSphere Template post-processor takes an artifact from the VMware-iso builder, built on ESXi (i.e. remote)
or an artifact from the [vSphere](/docs/post-processors/vsphere.html) post-processor and allows to mark a VM as a
template and leaving it in a path of choice.

## Example

An example is shown below, showing only the post-processor configuration:

``` json
{
   "type": "vsphere-template",
   "host": "vcenter.local",
   "insecure": true,
   "username": "root",
   "password": "secret",
   "datacenter": "mydatacenter",
   "folder": "/packer-templates/os/distro-7"
}
```

## Configuration

There are many configuration options available for the post-processor. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

Required:

-   `host` (string) - The vSphere host that contains the VM built by the vmware-iso.

-   `password` (string) - Password to use to authenticate to the vSphere endpoint.

-   `username` (string) - The username to use to authenticate to the vSphere endpoint.

Optional:

-   `datacenter` (string) - If you have more than one, you will need to specify which one the ESXi used.

-   `folder` (string) - Target path where the template will be created.

-   `insecure` (boolean) - If it's true skip verification of server certificate. Default is false  

## Using the vSphere Template with local builders

Once the [vSphere](/docs/post-processors/vsphere.html) takes an artifact from the VMware builder and uploads it
to a vSphere endpoint, you will likely want to mark that VM as template. Packer can do this for you automatically
using a sequence definition (a collection of post-processors that are treated as as single pipeline, see
[Post-Processors](/docs/templates/post-processors.html) for more information):

``` json
{
  "post-processors": [
    [
      {
        "type": "vsphere",
         ...
      },
      {
        "type": "vsphere-template",
         ...
      }
    ]
  ]
}
```

In the example above, the result of each builder is passed through the defined sequence of post-processors starting
with the `vsphere` post-processor which will upload the artifact to a vSphere endpoint. The resulting artifact is then
passed on to the `vsphere-template` post-processor which handles marking a VM as a template.
