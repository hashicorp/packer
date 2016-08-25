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

### Optional:

-   `azure_tags` (object of name/value strings) - the user can define up to 15 tags.  Tag names cannot exceed 512 
    characters, and tag values cannot exceed 256 characters.  Tags are applied to every resource deployed by a Packer
    build, i.e. Resource Group, VM, NIC, VNET, Public IP, KeyVault, etc.

-   `cloud_environment_name` (string) One of `Public`, `China`, `Germany`, or
    `USGovernment`. Defaults to `Public`. Long forms such as
    `USGovernmentCloud` and `AzureUSGovernmentCloud` are also supported.

-   `image_version` (string) Specify a specific version of an OS to boot from. Defaults to `latest`.  There may be a
     difference in versions available across regions due to image synchronization latency.  To ensure a consistent
     version across regions set this value to one that is available in all regions where you are deploying.

    CLI example `azure vm image list -l westus -p Canonical -o UbuntuServer -k 16.04.0-LTS`

-   `image_url` (string) Specify a custom VHD to use.  If this value is set, do not set image_publisher, image_offer,
     image_sku, or image_version.

-   `tenant_id` (string) The account identifier with which your `client_id` and `subscription_id` are associated. If not
     specified, `tenant_id` will be looked up using `subscription_id`.

-   `object_id` (string) Specify an OAuth Object ID to protect WinRM certificates
    created at runtime.  This variable is required when creating images based on
    Windows; this variable is not used by non-Windows builds.  See `Windows`
    behavior for `os_type`, below.

-   `os_type` (string) If either `Linux` or `Windows` is specified Packer will
    automatically configure authentication credentials for your machine. For
    `Linux` this configures an SSH authorized key. For `Windows` this
    configures your Tenant ID, Object ID, Key Vault Name, Key Vault Secret, and
    WinRM certificate URL.

-   `virtual_network_name` (string) Use a pre-existing virtual network for the VM.  This option enables private
    communication with the VM, no public IP address is **used** or **provisioned**.  This value should only be set if
    Packer is executed from a host on the same subnet / virtual network.

-   `virtual_network_resource_group_name` (string) If virtual_network_name is set, this value **may** also be set.  If
    virtual_network_name is set, and this value is not set the builder attempts to determine the resource group
    containing the virtual network.  If the resource group cannot be found, or it cannot be disambiguated, this value
    should be set.

-   `virtual_network_subnet_name` (string) If virtual_network_name is set, this value **may** also be set.  If
     virtual_network_name is set, and this value is not set the builder attempts to determine the subnet to use with
     the virtual network.  If the subnet cannot be found, or it cannot be disambiguated, this value should be set.

-   `vm_size` (string) Size of the VM used for building. This can be changed
    when you deploy a VM from your VHD. See
    [pricing](https://azure.microsoft.com/en-us/pricing/details/virtual-machines/) information. Defaults to `Standard_A1`.

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

    "os_type": "Linux",
    "image_publisher": "Canonical",
    "image_offer": "UbuntuServer",
    "image_sku": "14.04.4-LTS",
    
    "azure_tags": {
      "dept": "engineering"
    },

    "location": "West US",
    "vm_size": "Standard_A2"
}
```

## Implementation

\~&gt; **Warning!** This is an advanced topic. You do not need to understand the implementation to use the Azure
builder.

The Azure builder uses ARM
[templates](https://azure.microsoft.com/en-us/documentation/articles/resource-group-authoring-templates/) to deploy
resources.  ARM templates make it easy to express the what without having to express the how.

The Azure builder works under the assumption that it creates everything it needs to execute a build.  When the build has
completed it simply deletes the resource group to cleanup any runtime resources.  Resource groups are named using the
form `packer-Resource-Group-<random>`. The value `<random>` is a random value that is generated at every invocation of
packer.  The `<random>` value is re-used as much as possible when naming resources, so users can better identify and
group these transient resources when seen in their subscription.

 > The VHD is created on a user specified storage account, not a random one created at runtime.  When a virtual machine
 is captured the resulting VHD is stored on the same storage account as the source VHD.  The VHD created by Packer must
 persist after a build is complete, which is why the storage account is set by the user.

The basic steps for a build are:

 1. Create a resource group.
 1. Validate and deploy a VM template.
 1. Execute provision - defined by the user; typically shell commands.
 1. Power off and capture the VM.
 1. Delete the resource group.
 1. Delete the temporary VM's OS disk.

The templates used for a build are currently fixed in the code.  There is a template for Linux, Windows, and KeyVault.
The templates are themselves templated with place holders for names, passwords, SSH keys, certificates, etc.

### What's Randomized?

The Azure builder creates the following random values at runtime.

 * Administrator Password: a random 32-character value using the *password alphabet*.
 * Certificate: a 2,048-bit certificate used to secure WinRM communication.  The certificate is valid for 24-hours, which starts roughly at invocation time.
 * Certificate Password: a random 32-character value using the *password alphabet* used to protect the private key of the certificate.
 * Compute Name: a random 15-character name prefixed with pkrvm; the name of the VM.
 * Deployment Name: a random 15-character name prefixed with pkfdp; the name of the deployment.
 * KeyVault Name: a random 15-character name prefixed with pkrkv.
 * OS Disk Name: a random 15-character name prefixed with pkros.
 * Resource Group Name: a random 33-character name prefixed with packer-Resource-Group-.
 * SSH Key Pair: a 2,048-bit asymmetric key pair; can be overriden by the user.

The default alphabet used for random values is **0123456789bcdfghjklmnpqrstvwxyz**.  The alphabet was reduced (no
vowels) to prevent running afoul of Azure decency controls.

The password alphabet used for random values is **0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ**.

### Windows

The Windows implementation is very similar to the Linux build, with the exception that it deploys a template to
configure KeyVault. Packer communicates with a Windows VM using the WinRM protocol.  Windows VMs on Azure default to
using both password and certificate based authentication for WinRM.  The password is easily set via the VM ARM template,
but the certificate requires an intermediary. The intermediary for Azure is KeyVault.  The certificate is uploaded to a
new KeyVault provisioned in the same resource group as the VM.  When the Windows VM is deployed, it links to the
certificate in KeyVault, and Azure will ensure the certificate is injected as part of deployment.

The basic steps for a Windows build are:

  1. Create a resource group.
  1. Validate and deploy a KeyVault template.
  1. Validate and deploy a VM template.
  1. Execute provision - defined by the user; typically shell commands.
  1. Power off and capture the VM.
  1. Delete the resource group.
  1. Delete the temporary VM's OS disk.

A Windows build requires two templates and two deployments.  Unfortunately, the KeyVault and VM cannot be deployed at
the same time hence the need for two templates and deployments.  The time required to deploy a KeyVault template is
minimal, so overall impact is small.

 > The KeyVault certificate is protected using the object_id of the SPN.  This is why Windows builds require object_id,
 and an SPN.  The KeyVault is deleted when the resource group is deleted.

See the [examples/azure](https://github.com/mitchellh/packer/tree/master/examples/azure) folder in the packer project
for more examples.
