---
description: |
    In order to build VMs in Azure, Packer needs various configuration options.
    These options and how to obtain them are documented on this page.
layout: docs
page_title: 'Setup - Azure - Builders'
sidebar_current: 'docs-builders-azure-setup'
---

# Authorizing Packer Builds in Azure

In order to build VMs in Azure Packer needs 6 configuration options to be
specified:

-   `subscription_id` - UUID identifying your Azure subscription (where billing
    is handled)

-   `client_id` - UUID identifying the Active Directory service principal that
    will run your Packer builds

-   `client_secret` - service principal secret / password

-   `resource_group_name` - name of the resource group where your VHD(s) will
    be stored

-   `storage_account` - name of the storage account where your VHD(s) will be
    stored

-&gt; Behind the scenes Packer uses the OAuth protocol to authenticate against
Azure Active Directory and authorize requests to the Azure Service Management
API. These topics are unnecessarily complicated so we will try to ignore them
for the rest of this document.<br /><br />You do not need to understand how
OAuth works in order to use Packer with Azure, though the Active Directory
terms "service principal" and "role" will be useful for understanding Azure's
access policies.

In order to get all of the items above, you will need a username and password
for your Azure account.

## Device Login

Device login is an alternative way to authorize in Azure Packer. Device login
only requires you to know your Subscription ID. (Device login is only supported
for Linux based VMs.) Device login is intended for those who are first time
users, and just want to ''kick the tires.'' We recommend the SPN approach if
you intend to automate Packer.

> Device login is for **interactive** builds, and SPN is **automated** builds.

There are three pieces of information you must provide to enable device login
mode.

1.  SubscriptionID
2.  Resource Group - parent resource group that Packer uses to build an image.
3.  Storage Account - storage account where the image will be placed.

> Device login mode is enabled by not setting client\_id and client\_secret.

> Device login mode is for the Public and US Gov clouds only.

The device login flow asks that you open a web browser, navigate to
<a href="http://aka.ms/devicelogin" class="uri">http://aka.ms/devicelogin</a>,
and input the supplied code. This authorizes the Packer for Azure application
to act on your behalf. An OAuth token will be created, and stored in the user's
home directory (\~/.azure/packer/oauth-TenantID.json). This token is used if
the token file exists, and it is refreshed as necessary. The token file
prevents the need to continually execute the device login flow. Packer will ask
for two device login auth, one for service management endpoint and another for
accessing temp keyvault secrets that it creates.

## Managed identities for Azure resources

-&gt; Managed identities for Azure resources is the new name for the service
formerly known as Managed Service Identity (MSI).

Managed identities is an alternative way to authorize in Azure Packer. Managed
identities for Azure resources are automatically managed by Azure and enable
you to authenticate to services that support Azure AD authentication without
needing to insert credentials into your buildfile. Navigate to
<a href="https://docs.microsoft.com/en-gb/azure/active-directory/managed-identities-azure-resources/overview" 
class="uri">managed identities azure resources overview</a> to learn more about
this feature.

This feature will be used when no `subscription_id`, `client_id` or
`client_secret` is set in your buildfile.

## Install the Azure CLI

To get the credentials above, we will need to install the Azure CLI. Please
refer to Microsoft's official [installation
guide](https://azure.microsoft.com/en-us/documentation/articles/xplat-cli-install/).

-&gt; The guides below also use a tool called
[`jq`](https://stedolan.github.io/jq/) to simplify the output from the Azure
CLI, though this is optional. If you use homebrew you can simply
`brew install node jq`.

You can also use the Azure CLI in Docker. It also comes with `jq`
pre-installed:

``` shell
$ docker run -it microsoft/azure-cli
```

## Guided Setup

The Packer project includes a [setup
script](https://github.com/hashicorp/packer/blob/master/contrib/azure-setup.sh)
that can help you setup your account. It uses an interactive bash script to log
you into Azure, name your resources, and export your Packer configuration.

## Manual Setup

If you want more control or the script does not work for you, you can also use
the manual instructions below to setup your Azure account. You will need to
manually keep track of the various account identifiers, resource names, and
your service principal password.

### Identify Your Tenant and Subscription IDs

Login using the Azure CLI

``` shell
$ az login
# Note, we have launched a browser for you to login. For old experience with device code, use "az login --use-device-code"
```

Once you've completed logging in, you should get a JSON array like the one
below:

``` shell
[
  {
    "cloudName": "AzureCloud",
    "id": "$uuid",
    "isDefault": false,
    "name": "Pay-As-You-Go",
    "state": "Enabled",
    "tenantId": "$tenant_uuid",
    "user": {
      "name": "my_email@anywhere.com",
      "type": "user"
    }
  }
]
```

Get your account information

``` shell
$ az account list --output json | jq -r '.[].name'
$ az account set --subscription ACCOUNTNAME
$ az account show --output json | jq -r '.id'
```

-&gt; Throughout this document when you see a command pipe to `jq` you may
instead omit `--output json` and everything after it, but the output will be
more verbose. For example you can simply run `az account list` instead.

This will print out one line that look like this:

    4f562e88-8caf-421a-b4da-e3f6786c52ec

This is your `subscription_id`. Note it for later.

### Create a Resource Group

A [resource
group](https://azure.microsoft.com/en-us/documentation/articles/resource-group-overview/#resource-groups)
is used to organize related resources. Resource groups and storage accounts are
tied to a location. To see available locations, run:

``` shell
$ az account list-locations
$ LOCATION=xxx
$ GROUPNAME=xxx
# ...

$ az group create --name $GROUPNAME --location $LOCATION
```

Your storage account (below) will need to use the same `GROUPNAME` and
`LOCATION`.

### Create a Storage Account

We will need to create a storage account where your Packer artifacts will be
stored. We will create a `LRS` storage account which is the least expensive
price/GB at the time of writing.

``` shell
$ az storage account create \
  --name STORAGENAME
  --resource-group $GROUPNAME \
  --location $LOCATION \
  --sku Standard_LRS \
  --kind Storage
```

-&gt; `LRS` and `Standard_LRS` are meant as literal "LRS" or "Standard\_LRS"
and not as variables.

Make sure that `GROUPNAME` and `LOCATION` are the same as above. Also, ensure
that `GROUPNAME` is less than 24 characters long and contains only lowercase
letters and numbers.

### Create an Application

An application represents a way to authorize access to the Azure API. Note that
you will need to specify a URL for your application (this is intended to be
used for OAuth callbacks) but these do not actually need to be valid URLs.

First pick APPNAME, APPURL and PASSWORD:

``` shell
APPNAME=packer.test
APPURL=packer.test
PASSWORD=xxx
```

Password is your `client_secret` and can be anything you like. I recommend
using `openssl rand -base64 24`.

``` shell
$ az ad app create \
  --display-name $APPNAME \
  --identifier-uris $APPURL \
  --homepage $APPURL \
  --password $PASSWORD
```

### Create a Service Principal

You cannot directly grant permissions to an application. Instead, you create a
service principal and assign permissions to the service principal. To create a
service principal for use with Packer, run the below command specifying the
subscription. This will grant Packer the contributor role to the subscription.
The output of this command is your service principal credentials, save these in
a safe place as you will need these to configure Packer.

``` shell
az ad sp create-for-rbac -n "Packer" --role contributor \
                            --scopes /subscriptions/{SubID}
```

The service principal credentials.

``` shell
{
  "appId": "AppId",
  "displayName": "Packer",
  "name": "http://Packer",
  "password": "Password",
  "tenant": "TenantId"
}
```

There are a lot of pre-defined roles and you can define your own with more
granular permissions, though this is out of scope. You can see a list of
pre-configured roles via:

``` shell
$ az role definition list --output json | jq ".[] | {name:.roleName, description:.description}"
```

### Configuring Packer

Now (finally) everything has been setup in Azure and our service principal has
been created. You can use the output from creating your service principal in
your template.
