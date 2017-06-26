---
description: |
    In order to build VMs in Azure, Packer needs various configuration options.
    These options and how to obtain them are documented on this page.
layout: docs
page_title: 'Setup - Azure - Builders'
sidebar_current: 'docs-builders-azure-setup'
---

# Authorizing Packer Builds in Azure

In order to build VMs in Azure Packer needs 6 configuration options to be specified:

-   `subscription_id` - UUID identifying your Azure subscription (where billing is handled)

-   `client_id` - UUID identifying the Active Directory service principal that will run your Packer builds

-   `client_secret` - service principal secret / password

-   `object_id` - service principal object id (OSType = Windows Only)

-   `resource_group_name` - name of the resource group where your VHD(s) will be stored

-   `storage_account` - name of the storage account where your VHD(s) will be stored

-&gt; Behind the scenes Packer uses the OAuth protocol to authenticate against Azure Active Directory and authorize requests to the Azure Service Management API. These topics are unnecessarily complicated so we will try to ignore them for the rest of this document.<br /><br />You do not need to understand how OAuth works in order to use Packer with Azure, though the Active Directory terms "service principal" and "role" will be useful for understanding Azure's access policies.

In order to get all of the items above, you will need a username and password for your Azure account.

## Device Login

Device login is an alternative way to authorize in Azure Packer. Device login only requires you to know your
Subscription ID. (Device login is only supported for Linux based VMs.) Device login is intended for those who are first
time users, and just want to ''kick the tires.'' We recommend the SPN approach if you intend to automate Packer, or for
deploying Windows VMs.

> Device login is for **interactive** builds, and SPN is **automated** builds.

There are three pieces of information you must provide to enable device login mode.

1.  SubscriptionID
2.  Resource Group - parent resource group that Packer uses to build an image.
3.  Storage Account - storage account where the image will be placed.

> Device login mode is enabled by not setting client\_id and client\_secret.

> Device login mode is for the public cloud only, and Linux VMs only.

The device login flow asks that you open a web browser, navigate to <http://aka.ms/devicelogin>, and input the supplied
code. This authorizes the Packer for Azure application to act on your behalf. An OAuth token will be created, and stored
in the user's home directory (~/.azure/packer/oauth-TenantID.json). This token is used if the token file exists, and it
is refreshed as necessary. The token file prevents the need to continually execute the device login flow.

## Install the Azure CLI

To get the credentials above, we will need to install the Azure CLI. Please refer to Microsoft's official [installation guide](https://azure.microsoft.com/en-us/documentation/articles/xplat-cli-install/).

-&gt; The guides below also use a tool called [`jq`](https://stedolan.github.io/jq/) to simplify the output from the Azure CLI, though this is optional. If you use homebrew you can simply `brew install node jq`.

If you already have node.js installed you can use `npm` to install `azure-cli`:

``` shell
$ npm install -g azure-cli --no-progress
```

## Guided Setup

The Packer project includes a [setup script](https://github.com/hashicorp/packer/blob/master/contrib/azure-setup.sh) that can help you setup your account. It uses an interactive bash script to log you into Azure, name your resources, and export your Packer configuration.

## Manual Setup

If you want more control or the script does not work for you, you can also use the manual instructions below to setup your Azure account. You will need to manually keep track of the various account identifiers, resource names, and your service principal password.

### Identify Your Tenant and Subscription IDs

Login using the Azure CLI

``` shell
$ azure config mode arm
$ azure login -u USERNAME
```

Get your account information

``` shell
$ azure account list --json | jq -r '.[].name'
$ azure account set ACCOUNTNAME
$ azure account show --json | jq -r ".[] | .id"
```

-&gt; Throughout this document when you see a command pipe to `jq` you may instead omit `--json` and everything after it, but the output will be more verbose. For example you can simply run `azure account list` instead.

This will print out one line that look like this:

    4f562e88-8caf-421a-b4da-e3f6786c52ec

This is your `subscription_id`. Note it for later.

### Create a Resource Group

A [resource group](https://azure.microsoft.com/en-us/documentation/articles/resource-group-overview/#resource-groups) is used to organize related resources. Resource groups and storage accounts are tied to a location. To see available locations, run:

``` shell
$ azure location list
# ...

$ azure group create -n GROUPNAME -l LOCATION
```

Your storage account (below) will need to use the same `GROUPNAME` and `LOCATION`.

### Create a Storage Account

We will need to create a storage account where your Packer artifacts will be stored. We will create a `LRS` storage account which is the least expensive price/GB at the time of writing.

``` shell
$ azure storage account create \
  -g GROUPNAME \
  -l LOCATION \
  --sku-name LRS \
  --kind storage STORAGENAME
```

-&gt; `LRS` is meant as a literal "LRS" and not as a variable.

Make sure that `GROUPNAME` and `LOCATION` are the same as above.

### Create an Application

An application represents a way to authorize access to the Azure API. Note that you will need to specify a URL for your application (this is intended to be used for OAuth callbacks) but these do not actually need to be valid URLs.

``` shell
$ azure ad app create \
  -n APPNAME \
  -i APPURL \
  --home-page APPURL \
  -p PASSWORD
```

Password is your `client_secret` and can be anything you like. I recommend using `openssl rand -base64 24`.

### Create a Service Principal

You cannot directly grant permissions to an application. Instead, you create a service principal associated with the application and assign permissions to the service principal.

First, get the `APPID` for the application we just created.

``` shell
$ azure ad app list --json \
  | jq '.[] | select(.displayName | contains("APPNAME")) | .appId'
# ...

$ azure ad sp create --applicationId APPID
```

### Grant Permissions to Your Application

Finally, we will associate the proper permissions with our application's service principal. We're going to assign the `Owner` role to our Packer application and change the scope to manage our whole subscription. (The `Owner` role can be scoped to a specific resource group to further reduce the scope of the account.) This allows Packer to create temporary resource groups for each build.

``` shell
$ azure role assignment create \
  --spn APPURL \
  -o "Owner" \
  -c /subscriptions/SUBSCRIPTIONID
```

There are a lot of pre-defined roles and you can define your own with more granular permissions, though this is out of scope. You can see a list of pre-configured roles via:

``` shell
$ azure role list --json \
  | jq ".[] | {name:.Name, description:.Description}"
```

### Configuring Packer

Now (finally) everything has been setup in Azure. Let's get our configuration keys together:

Get `subscription_id`:

``` shell
$ azure account show --json \
  | jq ".[] | .id"
```

Get `client_id`

``` shell
$ azure ad app list --json \
  | jq '.[] | select(.displayName | contains("APPNAME")) | .appId'
```

Get `client_secret`

This cannot be retrieved. If you forgot this, you will have to delete and re-create your service principal and the associated permissions.

Get `object_id` (OSTYpe=Windows only)

``` shell
azure ad sp show -n CLIENT_ID
```

Get `resource_group_name`

``` shell
$ azure group list
```

Get `storage_account`

``` shell
$ azure storage account list
```
