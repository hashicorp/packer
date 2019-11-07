---
description: |
    Packer is able to create Azure VM images. To achieve this, Packer comes with
    multiple builders depending on the strategy you want to use to build the images.
layout: docs
page_title: 'Azure images - Builders'
sidebar_current: 'docs-builders-azure'
---

# Azure Virtual Machine Image Builders

Packer can create Azure virtual machine images through variety of ways 
depending on the strategy that you want to use for building the images. 
Packer supports the following builders for Azure images at the moment:

-   [azure-arm](/docs/builders/azure-arm.html) - Uses Azure Resource
    Manager (ARM) to launch a virtual machine (VM) from which a new image is
    captured after provisioning. If in doubt, use this builder; it is the
    easiest builder to get started with.

-   [azure-chroot](/docs/builders/azure-chroot.html) - Uses ARM to create
    a managed disk that is attached to an existing Azure VM that Packer is
    running on. Provisioning leverages [Chroot](https://en.wikipedia.org/wiki/Chroot)
    environment. After provisioning, the disk is detached an image is created
    from this disk. This is an **advanced builder and should not be used by
    newcomers**. However, it is also the fastest way to build a VM image in
    Azure.

-&gt; **Don't know which builder to use?** If in doubt, use the [azure-arm
builder](/docs/builders/azure-arm.html). It is much easier to use.

# Authentication for Azure

The Packer Azure builders provide a couple of ways to authenticate to Azure. The
following methods are available and are explained below:

-   Azure Active Directory interactive login. Interactive login is available
    for the Public and US Gov clouds only.
-   Azure Managed Identity
-   Azure Active Directory Service Principal

-&gt; **Don't know which authentication method to use?** Go with interactive
login to try out the builders. If you need packer to run automatically,
switch to using a Service Principal or Managed Identity.

No matter which method you choose, the identity you use will need the
appropriate permissions on Azure resources for Packer to operate. The minimal
set of permissions is highly dependent on the builder and its configuration.
An easy way to get started is to assign the identity the `Contributor` role at
the subscription level.

## Azure Active Directory interactive login

If your organization allows it, you can use a command line interactive login
method based on oAuth 'device code flow'. Packer will select this method when
you only specify a `subscription_id` in your builder configuration. When you
run Packer, it will ask you to visit a web site and input a code. This web site
will then authenticate you, satisfying any two-factor authentication policies
that your organization might have. The tokens are cached under the `.azure/packer`
directory in your home directory and will be reused if they are still valid
on subsequent runs.

## Azure Managed Identity

Azure provides the option to assign an identity to a virtual machine ([Azure
documentation](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/qs-configure-portal-windows-vm)). Packer can
use a system assigned identity for a VM where Packer is running to orchestrate
Azure API's. This is the default behavior and requires no configuration
properties to be set. It does, however, require that you run Packer on an
Azure VM.

To enable this method, [let Azure assign a system-assigned identity to your VM](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/qs-configure-portal-windows-vm).
Then, [grant your VM access to the appropriate resources](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/howto-assign-access-portal).
To get started, try assigning the `Contributor` role at the subscription level to
your VM. Then, when you discover your exact scenario, scope the permissions
appropriately or isolate Packer builds in a separate subscription.

##  Azure Active Directory Service Principal

Azure Active Directory models service accounts as 'Service Principal' (SP)
objects. An SP represents an application accessing your Azure resources. It
is identified by a client ID (aka application ID) and can use a password or a
certificate to authenticate. To use a Service Principal, specify the 
`subscription_id` and `client_id`, as well as either `client_secret`,
`client_cert_path` or `client_jwt`. Each of these last three represent a different
way to authenticate the SP to AAD:

-   `client_secret` - allows the user to provide a password/secret registered 
    for the AAD SP.
-   `client_cert_path` - allows usage of a certificate to be used to
    authenticate as the specified AAD SP.
-   `client_jwt` - For advanced scenario's where the used cannot provide Packer
    the full certificate, they can provide a JWT bearer token for client auth
    (RFC 7523, Sec. 2.2). These bearer tokens are created and signed using a
    certificate registered in AAD and have a user-chosen expiry time, limiting
    the validity of the token. This is also the underlying mechanism used to
    authenticate when using `client_cert_path`.

To create a service principal, you can follow [the Azure documentation on this
subject](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest).

