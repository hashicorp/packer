---
description: |
    The `cloudstack` Packer builder is able to create new templates for use with
    CloudStack. The builder takes either an ISO or an existing template as it's
    source, runs any provisioning necessary on the instance after launching it
    and then creates a new template from that instance.
layout: docs
page_title: CloudStack Builder
...

# CloudStack Builder

Type: `cloudstack`

The `cloudstack` Packer builder is able to create new templates for use with
[CloudStack](https://cloudstack.apache.org/). The builder takes either an ISO
or an existing template as it's source, runs any provisioning necessary on the
instance after launching it and then creates a new template from that instance.

The builder does *not* manage templates. Once a template is created, it is up
to you to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `api_url` (string) - The CloudStack API endpoint we will connect to.

-   `api_key` (string) - The API key used to sign all API requests.

-   `cidr_list` (array) - List of CIDR's that will have access to the new
    instance. This is needed in order for any provisioners to be able to
    connect to the instance. Usually this will be the NAT address of your
    current location. Only required when `use_local_ip_address` is `false`.

-   `instance_name` (string) - The name of the instance. Defaults to
    "packer-UUID" where UUID is dynamically generated.

-   `network` (string) - The name or ID of the network to connect the instance
    to.

-   `secret_key` (string) - The secret key used to sign all API requests.

-   `service_offering` (string) - The name or ID of the service offering used
    for the instance.

-   `soure_iso` (string) - The name or ID of an ISO that will be mounted before
    booting the instance. This option is mutual exclusive with `source_template`.

-   `source_template` (string) - The name or ID of the template used as base
    template for the instance. This option is mutual explusive with `source_iso`.

-   `template_name` (string) - The name of the new template. Defaults to
    "packer-{{timestamp}}" where timestamp will be the current time.

-   `template_display_text` (string) - The display text of the new template.
    Defaults to the `template_name`.

-   `template_os` (string) - The name or ID of the template OS for the new
    template that will be created.

-   `zone` (string) - The name or ID of the zone where the instance will be
    created.

### Optional:

-   `async_timeout` (int) - The time duration to wait for async calls to
    finish. Defaults to 30m.

-   `disk_offering` (string) - The name or ID of the disk offering used for the
    instance. This option is only available (and also required) when using
    `source_iso`.

-   `disk_size` (int) - The size (in GB) of the root disk of the new instance.
    This option is only available when using `source_template`.

-   `http_get_only` (boolean) - Some cloud providers only allow HTTP GET calls to
    their CloudStack API. If using such a provider, you need to set this to `true`
    in order for the provider to only make GET calls and no POST calls.

-   `hypervisor` (string) - The target hypervisor (e.g. `XenServer`, `KVM`) for
    the new template. This option is required when using `source_iso`.

-   `keypair` (string) - The name of the SSH key pair that will be used to
    access the instance. The SSH key pair is assumed to be already available
    within CloudStack.

-   `project` (string) - The name or ID of the project to deploy the instance to.

-   `public_ip_address` (string) - The public IP address or it's ID used for
    connecting any provisioners to. If not provided, a temporary public IP
    address will be associated and released during the Packer run.

-   `ssl_no_verify` (boolean) - Set to `true` to skip SSL verification. Defaults
    to `false`.

-   `template_featured` (boolean) - Set to `true` to indicate that the template
    is featured. Defaults to `false`.

-   `template_public` (boolean) - Set to `true` to indicate that the template is
    available for all accounts. Defaults to `false`.

-   `template_password_enabled` (boolean) - Set to `true` to indicate the template
    should be password enabled. Defaults to `false`.

-   `template_requires_hvm` (boolean) - Set to `true` to indicate the template
    requires hardware-assisted virtualization. Defaults to `false`.

-   `template_scalable` (boolean) - Set to `true` to indicate that the template
    contains tools to support dynamic scaling of VM cpu/memory. Defaults to `false`.

-   `user_data` (string) - User data to launch with the instance.

-   `use_local_ip_address` (boolean) - Set to `true` to indicate that the
    provisioners should connect to the local IP address of the instance.

## Basic Example

Here is a basic example.

``` {.javascript}
{
  "type": "cloudstack",
  "api_url": "https://cloudstack.company.com/client/api",
  "api_key": "YOUR_API_KEY",
  "secret_key": "YOUR_SECRET_KEY",

  "disk_offering": "Small - 20GB",
  "cidr_list": ["0.0.0.0/0"]
  "hypervisor": "KVM",
  "network": "management",
  "service_offering": "small",
  "source_iso": "CentOS-7.0-1406-x86_64-Minimal",
  "zone": "NL1",

  "template_name": "Centos7-x86_64-KVM-Packer",
  "template_display_text": "Centos7-x86_64 KVM Packer",
  "template_featured": true,
  "template_password_enabled": true,
  "template_scalable": true,
  "template_os": "Other PV (64-bit)"
}
```
