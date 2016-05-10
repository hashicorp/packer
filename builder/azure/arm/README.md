# packer-azure-arm

The ARM flavor of packer-azure utilizes the
[Azure Resource Manager APIs](https://msdn.microsoft.com/en-us/library/azure/dn790568.aspx).
Please see the
[overview](https://azure.microsoft.com/en-us/documentation/articles/resource-group-overview/)
for more information about ARM as well as the benefit of ARM.

## Device Login vs. Service Principal Name (SPN)

There are two ways to get started with packer-azure.  The simplest is device login, and only requires a Subscription ID.
Device login is only supported for Linux based VMs. The second is the use of an SPN.  We recommend the device login
approach for those who are first time users, and just want to ''kick the tires.''  We recommend the SPN approach if you
intend to automate Packer, or you are deploying Windows VMs.

## Device Login

A sample template for device login is show below.  There are three pieces of information
you must provide to enable device login mode.

 1. SubscriptionID
 1. Resource Group - parent resource group that Packer uses to build an image.
 1. Storage Account - storage account where the image will be placed.

> Device login mode is enabled by not setting client_id, client_secret, and tenant_id.

The device login flow asks that you open a web browser, navigate to http://aka.ms/devicelogin, and input the supplied
code.  This authorizes the Packer for Azure application to act on your behalf. An OAuth token will be created, and
stored in the user's home directory (~/.azure/packer/oauth-TenantID.json, and TenantID will be replaced with the actual
Tenant ID).  This token is used if it exists, and refreshed as necessary.

```json
{
    "variables": {
        "sid": "your_subscription_id",
        "rgn": "your_resource_group",
        "sa": "your_storage_account"
    },
    "builders": [
        {
            "type": "azure-arm",

            "subscription_id": "{{user `sid`}}",

            "resource_group_name": "{{user `rgn`}}",
            "storage_account": "{{user `sa`}}",

            "capture_container_name": "images",
            "capture_name_prefix": "packer",

            "os_type": "Linux",
            "image_publisher": "Canonical",
            "image_offer": "UbuntuServer",
            "image_sku": "14.04.3-LTS",

            "location": "South Central US",
            "vm_size": "Standard_A2"
        }
    ],
    "provisioners": [
        {
            "execute_command": "chmod +x {{ .Path }}; {{ .Vars }} sudo -E sh '{{ .Path }}'",
            "inline": [
                "apt-get update",
                "apt-get upgrade -y",

                "/usr/sbin/waagent -force -deprovision+user && export HISTSIZE=0 && sync"
            ],
            "inline_shebang": "/bin/sh -x",
            "type": "shell"
        }
    ]
}
```

## Service Principal Name

The ARM APIs use OAUTH to authenticate, and requires an SPN.  The following articles
are a good starting points for creating a new SPN.

 * [Automating Azure on your CI server using a Service Principal](http://blog.davidebbo.com/2014/12/azure-service-principal.html)
 * [Authenticating a service principal with Azure Resource Manager](https://azure.microsoft.com/en-us/documentation/articles/resource-group-authenticate-service-principal/)

There are three (four in the case of Windows) pieces of configuration you need to note
after creating an SPN.

 1. Client ID (aka Service Principal ID)
 1. Client Secret (aka Service Principal generated key)
 1. Client Tenant (aka Azure Active Directory tenant that owns the
    Service Principal)
 1. Object ID (Windows only) - a certificate is used to authenticate WinRM access, and the certificate is injected into
    the VM using Azure Key Vault.  Access to the key vault is protected by an ACL associated with the SPN's ObjectID.
    Linux does not need nor use a key vault, so there's no need to know the ObjectID.

You will also need the following.

 1. Subscription ID
 1. Resource Group
 1. Storage Account

Resource Group is where your storage account is located, and Storage
Account is where the created packer image will be stored.

The Service Principal has been tested with the following [permissions](https://azure.microsoft.com/en-us/documentation/articles/role-based-access-control-configure/).
Please review the document for the [built in roles](https://azure.microsoft.com/en-gb/documentation/articles/role-based-access-built-in-roles/)
for more details.

 * Owner

> NOTE: the Owner role is too powerful, and more explicit set of roles
> is TBD.  Issue #183 is tracking this work.  Permissions can be scoped to
> a specific resource group to further limit access.

### Sample Ubuntu

The following is a sample Packer template for use with the Packer
Azure for ARM builder.

```json
{
    "variables": {
        "cid": "your_client_id",
        "cst": "your_client_secret",
        "tid": "your_client_tenant",
        "sid": "your_subscription_id",

        "rgn": "your_resource_group",
        "sa": "your_storage_account"
    },
    "builders": [
        {
            "type": "azure-arm",

            "client_id": "{{user `cid`}}",
            "client_secret": "{{user `cst`}}",
            "subscription_id": "{{user `sid`}}",
            "tenant_id": "{{user `tid`}}",

            "resource_group_name": "{{user `rgn`}}",
            "storage_account": "{{user `sa`}}",

            "capture_container_name": "images",
            "capture_name_prefix": "packer",

            "os_type": "Linux",
            "image_publisher": "Canonical",
            "image_offer": "UbuntuServer",
            "image_sku": "14.04.3-LTS",

            "location": "South Central US",

            "vm_size": "Standard_A2"
        }
    ],
    "provisioners": [
        {
            "execute_command": "chmod +x {{ .Path }}; {{ .Vars }} sudo -E sh '{{ .Path }}'",
            "inline": [
                "apt-get update",
                "apt-get upgrade -y",

                "/usr/sbin/waagent -force -deprovision+user && export HISTSIZE=0 && sync"
            ],
            "inline_shebang": "/bin/sh -x",
            "type": "shell"
        }
    ]
}
```

Using the above template, Packer would be invoked as follows.

> NOTE: the following variables must be **changed** based on your
> subscription.  These values are just dummy values, but they match
> format of expected, e.g. if the value is a GUID the sample is a
> GUID.

```bat
packer build^
  -var cid="593c4dc4-9cd7-49af-9fe0-1ea5055ac1e4"^
  -var cst="GbzJfsfrVkqL/TLfZY8TXA=="^
  -var sid="ce323e74-56fc-4bd6-aa18-83b6dc262748"^
  -var tid="da3847b4-8e69-40bd-a2c2-41da6982c5e2"^
  -var rgn="My Resource Group"^
  -var sa="mystorageaccount"^
  c:\packer\ubuntu_14_LTS.json
```

Please see the
[config_sameples/arm](https://github.com/Azure/packer-azure/tree/master/config_examples)
directory for more examples of usage.
