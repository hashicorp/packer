---
description: |
    The cloudstack Packer builder is able to create new templates for use with
    CloudStack. The builder takes either an ISO or an existing template as it's
    source, runs any provisioning necessary on the instance after launching it and
    then creates a new template from that instance.
layout: docs
page_title: 'CloudStack - Builders'
sidebar_current: 'docs-builders-cloudstack'
---

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
    It can also be specified via environment variable `CLOUDSTACK_API_URL`,
    if set.

-   `api_key` (string) - The API key used to sign all API requests. It
    can also be specified via environment variable `CLOUDSTACK_API_KEY`,
    if set.

-   `network` (string) - The name or ID of the network to connect the instance
    to.

-   `secret_key` (string) - The secret key used to sign all API requests.
    It can also be specified via environment variable `CLOUDSTACK_SECRET_KEY`,
    if set.

-   `service_offering` (string) - The name or ID of the service offering used
    for the instance.

-   `source_iso` (string) - The name or ID of an ISO that will be mounted before
    booting the instance. This option is mutually exclusive with `source_template`.
    When using `source_iso`, both `disk_offering` and `hypervisor` are required.

-   `source_template` (string) - The name or ID of the template used as base
    template for the instance. This option is mutually exclusive with `source_iso`.

-   `template_os` (string) - The name or ID of the template OS for the new
    template that will be created.

-   `zone` (string) - The name or ID of the zone where the instance will be
    created.

### Optional:

-   `async_timeout` (number) - The time duration to wait for async calls to
    finish. Defaults to 30m.

-   `cidr_list` (array) - List of CIDR's that will have access to the new
    instance. This is needed in order for any provisioners to be able to
    connect to the instance. Defaults to `[ "0.0.0.0/0" ]`. Only required
    when `use_local_ip_address` is `false`.

-   `create_security_group` (boolean) - If `true` a temporary security group
    will be created which allows traffic towards the instance from the
    `cidr_list`. This option will be ignored if `security_groups` is also
    defined. Requires `expunge` set to `true`. Defaults to `false`.

-   `disk_offering` (string) - The name or ID of the disk offering used for the
    instance. This option is only available (and also required) when using
    `source_iso`.

-   `disk_size` (number) - The size (in GB) of the root disk of the new instance.
    This option is only available when using `source_template`.

-   `expunge` (boolean) - Set to `true` to expunge the instance when it is
    destroyed. Defaults to `false`.

-   `http_directory` (string) - Path to a directory to serve using an
    HTTP server. The files in this directory will be available over HTTP that
    will be requestable from the virtual machine. This is useful for hosting
    kickstart files and so on. By default this is "", which means no HTTP server
    will be started. The address and port of the HTTP server will be available
    as variables in `user_data`. This is covered in more detail below.

-   `http_get_only` (boolean) - Some cloud providers only allow HTTP GET calls to
    their CloudStack API. If using such a provider, you need to set this to `true`
    in order for the provider to only make GET calls and no POST calls.

-   `http_port_min` and `http_port_max` (number) - These are the minimum and
    maximum port to use for the HTTP server started to serve the
    `http_directory`. Because Packer often runs in parallel, Packer will choose
    a randomly available port in this range to run the HTTP server. If you want
    to force the HTTP server to be on one port, make this minimum and maximum
    port the same. By default the values are 8000 and 9000, respectively.

-   `hypervisor` (string) - The target hypervisor (e.g. `XenServer`, `KVM`) for
    the new template. This option is required when using `source_iso`.

-   `keypair` (string) - The name of the SSH key pair that will be used to
    access the instance. The SSH key pair is assumed to be already available
    within CloudStack.

-   `instance_name` (string) - The name of the instance. Defaults to
    "packer-UUID" where UUID is dynamically generated.

-   `project` (string) - The name or ID of the project to deploy the instance to.

-   `public_ip_address` (string) - The public IP address or it's ID used for
    connecting any provisioners to. If not provided, a temporary public IP
    address will be associated and released during the Packer run.

-   `security_groups` (array of strings) - A list of security group IDs or names
    to associate the instance with.

-   `ssh_agent_auth` (boolean) - If true, the local SSH agent will be used to
    authenticate connections to the source instance. No temporary keypair will
    be created, and the values of `ssh_password` and `ssh_private_key_file` will
    be ignored. To use this option with a key pair already configured in the source
    image, leave the `keypair` blank. To associate an existing key pair
    with the source instance, set the `keypair` field to the name of the key pair.

-   `ssl_no_verify` (boolean) - Set to `true` to skip SSL verification. Defaults
    to `false`.

-   `template_display_text` (string) - The display text of the new template.
    Defaults to the `template_name`.

-   `template_featured` (boolean) - Set to `true` to indicate that the template
    is featured. Defaults to `false`.

-   `template_name` (string) - The name of the new template. Defaults to
    "packer-{{timestamp}}" where timestamp will be the current time.

-   `template_public` (boolean) - Set to `true` to indicate that the template is
    available for all accounts. Defaults to `false`.

-   `template_password_enabled` (boolean) - Set to `true` to indicate the template
    should be password enabled. Defaults to `false`.

-   `template_requires_hvm` (boolean) - Set to `true` to indicate the template
    requires hardware-assisted virtualization. Defaults to `false`.

-   `template_scalable` (boolean) - Set to `true` to indicate that the template
    contains tools to support dynamic scaling of VM cpu/memory. Defaults to `false`.

-   `temporary_keypair_name` (string) - The name of the temporary SSH key pair
    to generate. By default, Packer generates a name that looks like
    `packer_<UUID>`, where &lt;UUID&gt; is a 36 character unique identifier.

-   `user_data` (string) - User data to launch with the instance. This is a
    [template engine](/docs/templates/engine.html) see _User Data_ bellow for more
    details.

-   `user_data_file` (string) - Path to a file that will be used for the user
    data when launching the instance. This file will be parsed as a
    [template engine](/docs/templates/engine.html) see _User Data_ bellow for more
    details.

-   `use_local_ip_address` (boolean) - Set to `true` to indicate that the
    provisioners should connect to the local IP address of the instance.

## User Data

The available variables are:

-  `HTTPIP` and `HTTPPort` - The IP and port, respectively of an HTTP server
    that is started serving the directory specified by the `http_directory`
    configuration parameter. If `http_directory` isn't specified, these will be
    blank.

## Basic Example

Here is a basic example.

``` json
{
  "type": "cloudstack",
  "api_url": "https://cloudstack.company.com/client/api",
  "api_key": "YOUR_API_KEY",
  "secret_key": "YOUR_SECRET_KEY",

  "disk_offering": "Small - 20GB",
  "hypervisor": "KVM",
  "network": "management",
  "service_offering": "small",
  "source_iso": "CentOS-7.0-1406-x86_64-Minimal",
  "zone": "NL1",

  "ssh_username": "root",

  "template_name": "Centos7-x86_64-KVM-Packer",
  "template_display_text": "Centos7-x86_64 KVM Packer",
  "template_featured": true,
  "template_password_enabled": true,
  "template_scalable": true,
  "template_os": "Other PV (64-bit)"
}
```
