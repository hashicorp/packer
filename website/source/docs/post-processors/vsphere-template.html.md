---
description: |
    The Packer vSphere Template post-processor takes an artifact from the VMware-iso builder -**only if remote ESXI is chosen**-
    and allows to mark a VM as a template and leaving it in a path of choice. 
layout: docs
page_title: 'vSphere Template - Post-Processors'
sidebar_current: 'docs-post-processors-vSphere-template'
---

# vSphere Template Post-Processor

Type: `vsphere-tpl`

The Packer vSphere template post-processor takes an artifact from the VMware-iso builder
allows to mark a VM as a template and leaving it in a path of choice.

## Example

An example is shown below, showing only the post-processor configuration:

``` json
{  
   "type": "vsphere-tpl",
   "host": "vcenter.local",
   "username": "root",
   "password": "secret",
   "datacenter": "murlock",
   "vm_name": "distro-7.3",
   "folder": "/packer-templates/os/distro-7"
}
```

## Configuration

There are many configuration options available for the post-processor. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

Required:

-   `host` (string) - The vSphere host that contains the VM built by the vmware-iso.
    
-   `insecure` (boolean) - If it's true Skip verification of server certificate. 
    Default is false    

-   `password` (string) - Password to use to authenticate to the
    vSphere endpoint.

-   `username` (string) - The username to use to authenticate to the
    vSphere endpoint.

-   `vm_name` (string) - The name of the VM once it is uploaded.

Optional:

-   `folder` (string) - Target path where the template will be created. 
    
-   `datacenter` (string) - If you have more than one, you will need to specified one.
