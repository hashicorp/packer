# packer-azure-arm

The ARM flavor of packer-azure utilizes the
[Azure Resource Manager APIs](https://msdn.microsoft.com/en-us/library/azure/dn790568.aspx).
Please see the
[overview](https://azure.microsoft.com/en-us/documentation/articles/resource-group-overview/)
for more information about ARM as well as the benefit of ARM.

## Getting Started

The ARM APIs use OAUTH to authenticate, so you must create a Service
Principal.  The following articles are a good starting points.

 * [Automating Azure on your CI server using a Service Principal](http://blog.davidebbo.com/2014/12/azure-service-principal.html)
 * [Authenticating a service principal with Azure Resource Manager](https://azure.microsoft.com/en-us/documentation/articles/resource-group-authenticate-service-principal/)

There are three pieces of configuration you will need as a result of
creating a Service Principal.

 1. Client ID (aka Service Principal ID)
 1. Client Secret (aka Service Principal generated key)
 1. Client Tenant (aka Azure Active Directory tenant that owns the
    Service Principal)

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
> is TBD.  Issue #183 is tracking this work.

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

            "capture_container_name": "images",
            "capture_name_prefix": "my_prefix",

            "image_publisher": "Canonical",
            "image_offer": "UbuntuServer",
            "image_sku": "14.04.3-LTS",

            "location": "South Central US",

            "resource_group_name": "{{user `rgn`}}",
            "storage_account": "{{user `sa`}}",

            "vm_size": "Standard_A1"
        }
    ],
    "provisioners": [
        {
            "execute_command": "chmod +x {{ .Path }}; {{ .Vars }} sudo -E sh '{{ .Path }}'",
            "inline": [
                "sudo apt-get update",
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
