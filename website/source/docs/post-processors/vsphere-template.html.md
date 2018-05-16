---
description: |
    The Packer vSphere Template post-processor takes an artifact from the VMware-iso builder built on ESXi (i.e. remote)
    and allows to mark a VM as a template and leaving it in a path of choice. 
layout: docs
page_title: 'vSphere Template - Post-Processors'
sidebar_current: 'docs-post-processors-vSphere-template'
---

# vSphere Template Post-Processor

Type: `vsphere-template`

The Packer vSphere template post-processor takes an artifact from the VMware-iso builder built on ESXi (i.e. remote) and
allows to mark a VM as a template and leaving it in a path of choice.

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
