---
description: |
    
layout: docs
page_title: Azure Resource Manager
...

# Azure Resource Manager Builder

Type: `azure-arm`

Packer supports building VHDs in [Azure Resource Manager](https://azure.microsoft.com/en-us/documentation/articles/resource-group-overview/). Azure provides new users a [$200 credit for the first 30 days](https://azure.microsoft.com/en-us/free/); after which you will incur costs for VMs built and stored using Packer.

Unlike most Packer builders, the artifact produced by the ARM builder is a VHD (virtual hard disk), not a full virtual machine image. This means you will need to [perform some additional steps](https://github.com/Azure/packer-azure/issues/201) in order to launch a VM from your build artifact.

Azure uses a combination of OAuth and Active Directory to authorize requests to the ARM API. Learn how to [authorize access to ARM](/docs/builders/azure-setup.html).

The documentation below references command output from the [Azure CLI](https://azure.microsoft.com/en-us/documentation/articles/xplat-cli-install/).

-> At this time packer supports building Linux virtual machines in Azure. Support for building Windows VMs is in progress and will be added in an upcoming release.

## Configuration Reference

The following configuration options are available for building Azure images. In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `client_id` (string) The Active Directory service principal associated with your builder.

-   `client_secret` (string) The password or secret for your service principal.

-   `resource_group_name` (string) Resource group under which the final artifact will be stored.

-   `storage_account` (string) Storage account under which the final artifact will be stored.

-   `subscription_id` (string) Subscription under which the build will be performed. **The service principal specified in `client_id` must have full access to this subscription.**

-   `tenant_id` (string) The account identifier with which your `client_id` and `subscription_id` are associated.

-   `capture_container_name` (string) Destination container name. Essentially the "folder" where your VHD will be organized in Azure.

-   `capture_name_prefix` (string) VHD prefix. The final artifacts will be named `PREFIX-osDisk.UUID` and `PREFIX-vmTemplate.UUID`.

-   `image_publisher` (string) PublisherName for your base image. See [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/) for details.

    CLI example `azure vm image list-publishers -l westus`

-   `image_offer` (string) Offer for your base image. See [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/) for details.

    CLI example `azure vm image list-offers -l westus -p Canonical`

-   `image_sku` (string) SKU for your base image. See [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/) for details.

    CLI example `azure vm image list-skus -l westus -p Canonical -o UbuntuServer`

-   `location` (string) Azure datacenter in which your VM will build.

    CLI example `azure location list`

-   `vm_size` (string) Size of the VM used for building. This can be changed when you deploy a VM from your VHD. See [pricing](https://azure.microsoft.com/en-us/pricing/details/virtual-machines/) information.

    CLI example `azure vm sizes -l westus`

## Basic Example

Here is a basic example for Azure.

``` {.javascript}
{
    "type": "azure-arm",

    "client_id": "fe354398-d7sf-4dc9-87fd-c432cd8a7e09",
    "client_secret": "keepitsecret&#*$",
    "resource_group_name": "packerdemo",
    "storage_account": "virtualmachines",
    "subscription_id": "44cae533-4247-4093-42cf-897ded6e7823",
    "tenant_id": "de39842a-caba-497e-a798-7896aea43218",

    "capture_container_name": "images",
    "capture_name_prefix": "packer",

    "image_publisher": "Canonical",
    "image_offer": "UbuntuServer",
    "image_sku": "14.04.3-LTS",

    "location": "West US",
    "vm_size": "Standard_A2"
}
```

See the [examples/azure](https://github.com/mitchellh/packer/tree/master/examples/azure) folder in the packer project for more examples.
