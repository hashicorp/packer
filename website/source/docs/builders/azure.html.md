---
description: 'Packer supports building VHDs in Azure Resource manager.'
layout: docs
page_title: 'Azure - Builders'
sidebar_current: 'docs-builders-azure'
---

# Azure Resource Manager Builder

Type: `azure-arm`

Packer supports building VHDs in [Azure Resource
Manager](https://azure.microsoft.com/en-us/documentation/articles/resource-group-overview/).
Azure provides new users a [$200 credit for the first 30
days](https://azure.microsoft.com/en-us/free/); after which you will incur
costs for VMs built and stored using Packer.

Unlike most Packer builders, the artifact produced by the ARM builder is a VHD
(virtual hard disk), not a full virtual machine image. This means you will need
to [perform some additional
steps](https://github.com/Azure/packer-azure/issues/201) in order to launch a
VM from your build artifact.

Azure uses a combination of OAuth and Active Directory to authorize requests to
the ARM API. Learn how to [authorize access to
ARM](/docs/builders/azure-setup.html).

The documentation below references command output from the [Azure
CLI](https://azure.microsoft.com/en-us/documentation/articles/xplat-cli-install/).

## Configuration Reference

The following configuration options are available for building Azure images. In
addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required options for authentication:
If you're running packer on an Azure VM with a [managed identity](/docs/builders/azure-setup.html#managed-identities-for-azure-resources)
you don't need to specify any additional configuration options.
If you would like to use interactive user authentication, you should specify
`subscription_id` only. Packer will use cached credentials or redirect you
to a website to log in.
If you want to use a [service principal](/docs/builders/azure-setup.html#create-a-service-principal)
you should specify `subscription_id`, `client_id` and one of `client_secret`,
`client_cert_path` or `client_jwt`.

-   `subscription_id` (string) Subscription under which the build will be
    performed. **The service principal specified in `client_id` must have full
    access to this subscription, unless build\_resource\_group\_name option is
    specified in which case it needs to have owner access to the existing
    resource group specified in build\_resource\_group\_name parameter.**

-   `client_id` (string) The Active Directory service principal associated with
    your builder.

-   `client_secret` (string) The password or secret for your service principal.

-   `client_cert_path` (string) The location of a PEM file containing a
    certificate and private key for service principal.

-   `client_jwt` (string) The bearer JWT assertion signed using a certificate
    associated with your service principal principal. See [Azure Active
    Directory docs](https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-certificate-credentials)
    for more information.

### Required:

-   `image_publisher` (string) PublisherName for your base image. See
    [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
    for details.

    CLI example `az vm image list-publishers --location westus`

-   `image_offer` (string) Offer for your base image. See
    [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
    for details.

    CLI example
    `az vm image list-offers --location westus --publisher Canonical`

-   `image_sku` (string) SKU for your base image. See
    [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
    for details.

    CLI example
    `az vm image list-skus --location westus --publisher Canonical --offer UbuntuServer`

#### VHD or Managed Image

The Azure builder can create either a VHD, or a managed image. If you are
creating a VHD, you **must** start with a VHD. Likewise, if you want to create
a managed image you **must** start with a managed image. When creating a VHD
the following options are required.

-   `capture_container_name` (string) Destination container name. Essentially
    the "directory" where your VHD will be organized in Azure. The captured
    VHD's URL will be
    `https://<storage_account>.blob.core.windows.net/system/Microsoft.Compute/Images/<capture_container_name>/<capture_name_prefix>.xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.vhd`.

-   `capture_name_prefix` (string) VHD prefix. The final artifacts will be
    named `PREFIX-osDisk.UUID` and `PREFIX-vmTemplate.UUID`.

-   `resource_group_name` (string) Resource group under which the final
    artifact will be stored.

-   `storage_account` (string) Storage account under which the final artifact
    will be stored.

When creating a managed image the following options are required.

-   `managed_image_name` (string) Specify the managed image name where the
    result of the Packer build will be saved. The image name must not exist
    ahead of time, and will not be overwritten. If this value is set, the value
    `managed_image_resource_group_name` must also be set. See
    [documentation](https://docs.microsoft.com/en-us/azure/storage/storage-managed-disks-overview#images)
    to learn more about managed images.

-   `managed_image_resource_group_name` (string) Specify the managed image
    resource group name where the result of the Packer build will be saved. The
    resource group must already exist. If this value is set, the value
    `managed_image_name` must also be set. See
    [documentation](https://docs.microsoft.com/en-us/azure/storage/storage-managed-disks-overview#images)
    to learn more about managed images.


Managed images can optionally be published to [Shared Image Gallery](https://azure.microsoft.com/en-us/blog/announcing-the-public-preview-of-shared-image-gallery/)
as Shared Gallery Image version. Shared Image Gallery **only** works with Managed Images. **A VHD cannot be published to
a Shared Image Gallery**. When publishing to a Shared Image Gallery the following options are required.

- `shared_image_gallery_destination` (object) The name of the Shared Image Gallery under which the managed image will be published as Shared Gallery Image version.

Following is an example.

<!-- -->

    "shared_image_gallery_destination": {
        "resource_group": "ResourceGroup",
        "gallery_name": "GalleryName",
        "image_name": "ImageName",
        "image_version": "1.0.0",
        "replication_regions": ["regionA", "regionB", "regionC"]
    }
    "managed_image_name": "TargetImageName",
    "managed_image_resource_group_name": "TargetResourceGroup"

#### Resource Group Usage

The Azure builder can either provision resources into a new resource group that
it controls (default) or an existing one. The advantage of using a packer
defined resource group is that failed resource cleanup is easier because you
can simply remove the entire resource group, however this means that the
provided credentials must have permission to create and remove resource groups.
By using an existing resource group you can scope the provided credentials to
just this group, however failed builds are more likely to leave unused
artifacts.

To have packer create a resource group you **must** provide:

-   `location` (string) Azure datacenter in which your VM will build.

    CLI example `az account list-locations`

and optionally:

-   `temp_resource_group_name` (string) name assigned to the temporary resource
    group created during the build. If this value is not set, a random value
    will be assigned. This resource group is deleted at the end of the build.

To use an existing resource group you **must** provide:

-   `build_resource_group_name` (string) - Specify an existing resource group
    to run the build in.

Providing `temp_resource_group_name` or `location` in combination with
`build_resource_group_name` is not allowed.

### Optional:

-   `azure_tags` (object of name/value strings) - the user can define up to 15
    tags. Tag names cannot exceed 512 characters, and tag values cannot exceed
    256 characters. Tags are applied to every resource deployed by a Packer
    build, i.e. Resource Group, VM, NIC, VNET, Public IP, KeyVault, etc.

-   `cloud_environment_name` (string) One of `Public`, `China`, `Germany`, or
    `USGovernment`. Defaults to `Public`. Long forms such as
    `USGovernmentCloud` and `AzureUSGovernmentCloud` are also supported.

-   `custom_data_file` (string) Specify a file containing custom data to inject
    into the cloud-init process. The contents of the file are read and injected
    into the ARM template. The custom data will be passed to cloud-init for
    processing at the time of provisioning. See
    [documentation](http://cloudinit.readthedocs.io/en/latest/topics/examples.html)
    to learn more about custom data, and how it can be used to influence the
    provisioning process.

-   `custom_managed_image_name` (string) Specify the source managed image's
    name to use. If this value is set, do not set image\_publisher,
    image\_offer, image\_sku, or image\_version. If this value is set, the
    value `custom_managed_image_resource_group_name` must also be set. See
    [documentation](https://docs.microsoft.com/en-us/azure/storage/storage-managed-disks-overview#images)
    to learn more about managed images.

-   `custom_managed_image_resource_group_name` (string) Specify the source
    managed image's resource group used to use. If this value is set, do not
    set image\_publisher, image\_offer, image\_sku, or image\_version. If this
    value is set, the value `custom_managed_image_name` must also be set. See
    [documentation](https://docs.microsoft.com/en-us/azure/storage/storage-managed-disks-overview#images)
    to learn more about managed images.

-   `image_version` (string) Specify a specific version of an OS to boot from.
    Defaults to `latest`. There may be a difference in versions available
    across regions due to image synchronization latency. To ensure a consistent
    version across regions set this value to one that is available in all
    regions where you are deploying.

    CLI example
    `az vm image list --location westus --publisher Canonical --offer UbuntuServer --sku 16.04.0-LTS --all`

-   `image_url` (string) Specify a custom VHD to use. If this value is set, do
    not set image\_publisher, image\_offer, image\_sku, or image\_version.

-   `managed_image_storage_account_type` (string) Specify the storage account
    type for a managed image. Valid values are Standard\_LRS and Premium\_LRS.
    The default is Standard\_LRS.

-   `os_disk_size_gb` (number) Specify the size of the OS disk in GB
    (gigabytes). Values of zero or less than zero are ignored.

-   `disk_caching_type` (string) Specify the disk caching type. Valid values
    are None, ReadOnly, and ReadWrite. The default value is ReadWrite.

-   `disk_additional_size` (array of integers) - The size(s) of any additional
    hard disks for the VM in gigabytes. If this is not specified then the VM
    will only contain an OS disk. The number of additional disks and maximum
    size of a disk depends on the configuration of your VM. See
    [Windows](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/about-disks-and-vhds)
    or
    [Linux](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/about-disks-and-vhds)
    for more information.

    For VHD builds the final artifacts will be named
    `PREFIX-dataDisk-<n>.UUID.vhd` and stored in the specified capture
    container along side the OS disk. The additional disks are included in the
    deployment template `PREFIX-vmTemplate.UUID`.

    For Managed build the final artifacts are included in the managed image.
    The additional disk will have the same storage account type as the OS disk,
    as specified with the `managed_image_storage_account_type` setting.

-   `os_type` (string) If either `Linux` or `Windows` is specified Packer will
    automatically configure authentication credentials for the provisioned
    machine. For `Linux` this configures an SSH authorized key. For `Windows`
    this configures a WinRM certificate.

-   `plan_info` (object) - Used for creating images from Marketplace images.
    Please refer to [Deploy an image with Marketplace
    terms](https://aka.ms/azuremarketplaceapideployment) for more details. Not
    all Marketplace images support programmatic deployment, and support is
    controlled by the image publisher.

    An example plan\_info object is defined below.

    ``` json
    {
       "plan_info": {
           "plan_name": "rabbitmq",
           "plan_product": "rabbitmq",
           "plan_publisher": "bitnami"
       }
    }
    ```

    `plan_name` (string) - The plan name, required. `plan_product` (string) -
    The plan product, required. `plan_publisher` (string) - The plan publisher,
    required. `plan_promotion_code` (string) - Some images accept a promotion
    code, optional.

    Images created from the Marketplace with `plan_info` **must** specify
    `plan_info` whenever the image is deployed. The builder automatically adds
    tags to the image to ensure this information is not lost. The following
    tags are added.

    1.  PlanName
    2.  PlanProduct
    3.  PlanPublisher
    4.  PlanPromotionCode

-   `shared_image_gallery` (object) Use a [Shared Gallery
    image](https://azure.microsoft.com/en-us/blog/announcing-the-public-preview-of-shared-image-gallery/)
    as the source for this build. *VHD targets are incompatible with this build
    type* - the target must be a *Managed Image*.

<!-- -->

    "shared_image_gallery": {
        "subscription": "00000000-0000-0000-0000-00000000000",
        "resource_group": "ResourceGroup",
        "gallery_name": "GalleryName",
        "image_name": "ImageName",
        "image_version": "1.0.0"
    }
    "managed_image_name": "TargetImageName",
    "managed_image_resource_group_name": "TargetResourceGroup"

-   `shared_image_gallery_timeout` (time.Duration) How long to wait for an image
    to be published to the shared image gallery before timing out. If your
    Packer build is failing on the Publishing to Shared Image Gallery step
    with the error `Original Error: context deadline exceeded`, but the image
    is present when you check your Azure dashboard, then you probably need to
    increase this timeout from its default of "60m" (valid time units include
    `s` for seconds, `m` for minutes, and `h` for hours.)

-   `temp_compute_name` (string) temporary name assigned to the VM. If this
    value is not set, a random value will be assigned. Knowing the resource
    group and VM name allows one to execute commands to update the VM during a
    Packer build, e.g. attach a resource disk to the VM.

-   `tenant_id` (string) The account identifier with which your `client_id` and
    `subscription_id` are associated. If not specified, `tenant_id` will be
    looked up using `subscription_id`.

-   `private_virtual_network_with_public_ip` (boolean) This value allows you to
    set a `virtual_network_name` and obtain a public IP. If this value is not
    set and `virtual_network_name` is defined Packer is only allowed to be
    executed from a host on the same subnet / virtual network.

-   `virtual_network_name` (string) Use a pre-existing virtual network for the
    VM. This option enables private communication with the VM, no public IP
    address is **used** or **provisioned** (unless you set
    `private_virtual_network_with_public_ip`).

-   `virtual_network_resource_group_name` (string) If virtual\_network\_name is
    set, this value **may** also be set. If virtual\_network\_name is set, and
    this value is not set the builder attempts to determine the resource group
    containing the virtual network. If the resource group cannot be found, or
    it cannot be disambiguated, this value should be set.

-   `allowed_inbound_ip_addresses` (array of strings) list of IP addresses and
    CIDR blocks that should be allowed access to the VM. If provided, an Azure
    Network Security Group will be created with corresponding rules and be bound 
    to the NIC attached to the VM. 

-   `virtual_network_subnet_name` (string) If virtual\_network\_name is set,
    this value **may** also be set. If virtual\_network\_name is set, and this
    value is not set the builder attempts to determine the subnet to use with
    the virtual network. If the subnet cannot be found, or it cannot be
    disambiguated, this value should be set.

-   `vm_size` (string) Size of the VM used for building. This can be changed
    when you deploy a VM from your VHD. See
    [pricing](https://azure.microsoft.com/en-us/pricing/details/virtual-machines/)
    information. Defaults to `Standard_A1`.

    CLI example `az vm list-sizes --location westus`

-   `async_resourcegroup_delete` (boolean) If you want packer to delete the
    temporary resource group asynchronously set this value. It's a boolean
    value and defaults to false. **Important** Setting this true means that
    your builds are faster, however any failed deletes are not reported.

-   `managed_image_os_disk_snapshot_name` (string) If
    managed\_image\_os\_disk\_snapshot\_name is set, a snapshot of the OS disk
    is created with the same name as this value before the VM is captured.

-   `managed_image_data_disk_snapshot_prefix` (string) If
    managed\_image\_data\_disk\_snapshot\_prefix is set, snapshot of the data
    disk(s) is created with the same prefix as this value before the VM is
    captured.

-   `managed_image_zone_resilient` (bool) Store the image in zone-resilient storage. You need to create it
    in a region that supports [availability zones](https://docs.microsoft.com/en-us/azure/availability-zones/az-overview).

## Basic Example

Here is a basic example for Azure.

``` json
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

## Deprovision

Azure VMs should be deprovisioned at the end of every build. For Windows this
means executing sysprep, and for Linux this means executing the waagent
deprovision process.

Please refer to the Azure
[examples](https://github.com/hashicorp/packer/tree/master/examples/azure) for
complete examples showing the deprovision process.

### Windows

The following provisioner snippet shows how to sysprep a Windows VM.
Deprovision should be the last operation executed by a build. The code below
will wait for sysprep to write the image status in the registry and will exit
after that. The possible states, in case you want to wait for another state,
[are documented
here](https://technet.microsoft.com/en-us/library/hh824815.aspx)

``` json
{
    "provisioners": [
    {
        "type": "powershell",
        "inline": [
            " # NOTE: the following *3* lines are only needed if the you have installed the Guest Agent.",
            "  while ((Get-Service RdAgent).Status -ne 'Running') { Start-Sleep -s 5 }",
            "  while ((Get-Service WindowsAzureTelemetryService).Status -ne 'Running') { Start-Sleep -s 5 }",
            "  while ((Get-Service WindowsAzureGuestAgent).Status -ne 'Running') { Start-Sleep -s 5 }",

            "& $env:SystemRoot\\System32\\Sysprep\\Sysprep.exe /oobe /generalize /quiet /quit",
            "while($true) { $imageState = Get-ItemProperty HKLM:\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Setup\\State | Select ImageState; if($imageState.ImageState -ne 'IMAGE_STATE_GENERALIZE_RESEAL_TO_OOBE') { Write-Output $imageState.ImageState; Start-Sleep -s 10  } else { break } }"
        ]
    }
  ]
}
```

The Windows Guest Agent participates in the Sysprep process. The agent must be
fully installed before the VM can be sysprep'ed. To ensure this is true all
agent services must be running before executing sysprep.exe. The above JSON
snippet shows one way to do this in the PowerShell provisioner. This snippet is
**only** required if the VM is configured to install the agent, which is the
default. To learn more about disabling the Windows Guest Agent please see
[Install the VM
Agent](https://docs.microsoft.com/en-us/azure/virtual-machines/extensions/agent-windows#install-the-vm-agent).

### Linux

The following provisioner snippet shows how to deprovision a Linux VM.
Deprovision should be the last operation executed by a build.

``` json
{
 "provisioners": [
   {
     "execute_command": "chmod +x {{ .Path }}; {{ .Vars }} sudo -E sh '{{ .Path }}'",
     "inline": [
       "/usr/sbin/waagent -force -deprovision+user && export HISTSIZE=0 && sync"
     ],
     "inline_shebang": "/bin/sh -x",
     "type": "shell"
   }
 ]
}
```

To learn more about the Linux deprovision process please see WALinuxAgent's
[README](https://github.com/Azure/WALinuxAgent/blob/master/README.md).

#### skip\_clean

Customers have reported issues with the deprovision process where the builder
hangs. The error message is similar to the following.

    Build 'azure-arm' errored: Retryable error: Error removing temporary script at /tmp/script_9899.sh: ssh: handshake failed: EOF

One solution is to set skip\_clean to true in the provisioner. This prevents
Packer from cleaning up any helper scripts uploaded to the VM during the build.

## Defaults

The Azure builder attempts to pick default values that provide for a just works
experience. These values can be changed by the user to more suitable values.

-   The default user name is packer not root as in other builders. Most distros
    on Azure do not allow root to SSH to a VM hence the need for a non-root
    default user. Set the ssh\_username option to override the default value.
-   The default VM size is Standard\_A1. Set the vm\_size option to override
    the default value.
-   The default image version is latest. Set the image\_version option to
    override the default value.
-   By default a temporary resource group will be created and destroyed as part
    of the build. If you do not have permissions to do so, use
    `build_resource_group_name` to specify an existing resource group to run
    the build in.

## Implementation

\~&gt; **Warning!** This is an advanced topic. You do not need to understand
the implementation to use the Azure builder.

The Azure builder uses ARM
[templates](https://azure.microsoft.com/en-us/documentation/articles/resource-group-authoring-templates/)
to deploy resources. ARM templates allow you to express the what without having
to express the how.

The Azure builder works under the assumption that it creates everything it
needs to execute a build. When the build has completed it simply deletes the
resource group to cleanup any runtime resources. Resource groups are named
using the form `packer-Resource-Group-<random>`. The value `<random>` is a
random value that is generated at every invocation of packer. The `<random>`
value is re-used as much as possible when naming resources, so users can better
identify and group these transient resources when seen in their subscription.

> The VHD is created on a user specified storage account, not a random one
> created at runtime. When a virtual machine is captured the resulting VHD is
> stored on the same storage account as the source VHD. The VHD created by
> Packer must persist after a build is complete, which is why the storage
> account is set by the user.

The basic steps for a build are:

1.  Create a resource group.
2.  Validate and deploy a VM template.
3.  Execute provision - defined by the user; typically shell commands.
4.  Power off and capture the VM.
5.  Delete the resource group.
6.  Delete the temporary VM's OS disk.

The templates used for a build are currently fixed in the code. There is a
template for Linux, Windows, and KeyVault. The templates are themselves
templated with place holders for names, passwords, SSH keys, certificates, etc.

### What's Randomized?

The Azure builder creates the following random values at runtime.

-   Administrator Password: a random 32-character value using the *password
    alphabet*.
-   Certificate: a 2,048-bit certificate used to secure WinRM communication.
    The certificate is valid for 24-hours, which starts roughly at invocation
    time.
-   Certificate Password: a random 32-character value using the *password
    alphabet* used to protect the private key of the certificate.
-   Compute Name: a random 15-character name prefixed with pkrvm; the name of
    the VM.
-   Deployment Name: a random 15-character name prefixed with pkfdp; the name
    of the deployment.
-   KeyVault Name: a random 15-character name prefixed with pkrkv.
-   NIC Name: a random 15-character name prefixed with pkrni.
-   Public IP Name: a random 15-character name prefixed with pkrip.
-   OS Disk Name: a random 15-character name prefixed with pkros.
-   Resource Group Name: a random 33-character name prefixed with
    packer-Resource-Group-.
-   Subnet Name: a random 15-character name prefixed with pkrsn.
-   SSH Key Pair: a 2,048-bit asymmetric key pair; can be overridden by the
    user.
-   Virtual Network Name: a random 15-character name prefixed with pkrvn.

The default alphabet used for random values is
**0123456789bcdfghjklmnpqrstvwxyz**. The alphabet was reduced (no vowels) to
prevent running afoul of Azure decency controls.

The password alphabet used for random values is
**0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ**.

### Windows

The Windows implementation is very similar to the Linux build, with the
exception that it deploys a template to configure KeyVault. Packer communicates
with a Windows VM using the WinRM protocol. Windows VMs on Azure default to
using both password and certificate based authentication for WinRM. The
password is easily set via the VM ARM template, but the certificate requires an
intermediary. The intermediary for Azure is KeyVault. The certificate is
uploaded to a new KeyVault provisioned in the same resource group as the VM.
When the Windows VM is deployed, it links to the certificate in KeyVault, and
Azure will ensure the certificate is injected as part of deployment.

The basic steps for a Windows build are:

1.  Create a resource group.
2.  Validate and deploy a KeyVault template.
3.  Validate and deploy a VM template.
4.  Execute provision - defined by the user; typically shell commands.
5.  Power off and capture the VM.
6.  Delete the resource group.
7.  Delete the temporary VM's OS disk.

A Windows build requires two templates and two deployments. Unfortunately, the
KeyVault and VM cannot be deployed at the same time hence the need for two
templates and deployments. The time required to deploy a KeyVault template is
minimal, so overall impact is small.

See the
[examples/azure](https://github.com/hashicorp/packer/tree/master/examples/azure)
folder in the packer project for more examples.
