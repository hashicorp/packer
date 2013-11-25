---
layout: "docs"
page_title: "vSphere Post-Processor"
---

# vSphere Post-Processor

Type: `vsphere-upload`

The vSphere post-processor takes an artifact from the VMware builder
and uploads it to a vSphere endpoint.

## Configuration

There are many configuration options available for the post-processor. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

Required:

* `datacenter` (string) - The name of the datacenter within vSphere to
  add the VM to.

* `datastore` (string) - The name of the datastore to store this VM.

* `host` (string) - The vSphere host that will be contacted to perform
  the VM upload.

* `password` (string) - Password to use to authenticate to the vSphere
  endpoint.

* `path_to_resource_pool` (string) - The path within the resource pool to
  store the VM.

* `username` (string) - The username to use to authenticate to the vSphere
  endpoint.

* `vm_folder` (string) - The folder within the datastore to store the VM.

* `vm_name` (string) - The name of the VM once it is uploaded.

* `vm_network` (string) - The name of the VM network this VM will be
  added to.

Optional:

* `insecure` (bool) - Whether or not the connection to vSphere can be done
  over an insecure connection. By default this is false.
