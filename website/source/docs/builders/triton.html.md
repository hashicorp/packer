---
description: |
    The `triton` Packer builder is able to create new images for use with Triton. These images can be used with both the Joyent public cloud (which is powered by Triton) as well with private Triton installations. This builder uses the Triton Cloud API to create images. The builder creates and launches a temporary VM based on a specified source image, runs any provisioning necessary, uses the Triton "VM to image" functionality to create a reusable image and finally destroys the temporary VM. This reusable image can then be used to launch new VM's.
page_title: Triton Builder
...

# Triton Builder

Type: `triton`

The `triton` Packer builder is able to create new images for use with Triton. These images can be used with both the [Joyent public cloud](https://www.joyent.com/) (which is powered by Triton) as well with private [Triton](https://github.com/joyent/triton) installations. This builder uses the Triton Cloud API to create images. The builder creates and launches a temporary VM based on a specified source image, runs any provisioning necessary, uses the Triton "VM to image" functionality to create a reusable image and finally destroys the temporary VM. This reusable image can then be used to launch new VM's.

The builder does *not* manage images. Once it creates an image, it is up to you to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are segmented below into two categories: required and optional parameters.

In addition to the options listed here, a [communicator](/docs/templates/communicator.html) can be configured for this builder.

### Required:

-   `triton_account` (string) - The username of the Triton account to use when using the Triton Cloud API.
-   `triton_key_id` (string) - The fingerprint of the public key of the SSH key pair to use for authentication with the Triton Cloud API.
-   `triton_key_material` (string) - Path to the file in which the private key of `triton_key_id` is stored. For example `~/.ssh/id_rsa`.
 
 -   `source_machine_image` (string) - The UUID of the image to base the new image on. On the Joyent public cloud this could for example be `70e3ae72-96b6-11e6-9056-9737fd4d0764` for version 16.3.1 of the 64bit SmartOS base image.
-   `source_machine_package` (string) - The Triton package to use while building the image. Does not affect (and does not have to be the same) as the package which will be used for a VM instance running this image. On the Joyent public cloud this could for example be `g3-standard-0.5-smartos`.

-   `image_name` (string) - The name the finished image in Triton will be assigned. Maximum 512 characters but should in practice be much shorter (think between 5 and 20 characters). For example `postgresql-95-server` for an image used as a PostgreSQL 9.5 server.
-   `image_version` (string) - The version string for this image. Maximum 128 characters. Any string will do but a format of `Major.Minor.Patch` is strongly advised by Joyent. See [Semantic Versioning](http://semver.org/) for more information on the `Major.Minor.Patch` versioning format.

### Optional:

-   `triton_url` (string) - The URL of the Triton cloud API to use. If omitted it will default to the URL of the Joyent Public cloud. If you are using your own private Triton installation you will have to supply the URL of the cloud API of your own Triton installation.

-   `source_machine_firewall_enabled` (boolean) - Whether or not the firewall of the VM used to create an image of is enabled. The Triton firewall only filters inbound traffic to the VM. For the Joyent public cloud and private Triton installations SSH traffic is always allowed by default. All outbound traffic is always allowed. Currently this builder does not provide an interface to add specific firewall rules. The default is `false`.
-   `source_machine_metadata` (object of key/value strings) - Triton metadata applied to the VM used to create the image. Metadata can be used to pass configuration information to the VM without the need for networking. See [Using the metadata API](https://docs.joyent.com/private-cloud/instances/using-mdata) in the Joyent documentation for more information. This can for example be used to set the `user-script` metadata key to have Triton start a user supplied script after the VM has booted.
-   `source_machine_name` (string) - Name of the VM used for building the image. Does not affect (and does not have to be the same) as the name for a VM instance running this image. Maximum 512 characters but should in practice be much shorter (think between 5 and 20 characters). For example `mysql-64-server-image-builder`. When omitted defaults to `packer-builder-[image_name]`.
-   `source_machine_networks` (array of strings) - The UUID's of Triton networks added to the source machine used for creating the image. For example if any of the provisioners which are run need Internet access you will need to add the UUID's of the appropriate networks here. 
-   `source_machine_tags` (object of key/value strings) - Tags applied to the VM used to create the image.
-   `ssh_agent_auth` (boolean) - If true, the local SSH agent will be used to authenticate connections to the source VM. By default this value is `false` and the values of `triton_key_id` and `triton_key_material` will also be used for connecting to the VM.

-   `image_acls` (array of strings) - The UUID's of the users which will have access to this image. When omitted only the owner (the Triton user whose credentials are used) will have access to the image.
-   `image_description` (string) - Description of the image. Maximum 512 characters.
-   `image_eula_url` (string) - URL of the End User License Agreement (EULA) for the image. Maximum 128 characters.
-   `image_homepage` (string) - URL of the homepage where users can find information about the image. Maximum 128 characters.
-   `image_tags` (object of key/value strings) - Tag applied to the image.

## Basic Example

Below is a minimal example to create an image on the Joyent public cloud:

``` {.javascript}
"builders": [{
  "type": "triton",
  "triton_account": "triton_username",
  "triton_key_id": "6b:95:03:3d:d3:6e:52:69:01:96:1a:46:4a:8d:c1:7e",
  "triton_key_material": "${file("~/.ssh/id_rsa")}",
  "source_machine_name": "image-builder",
  "source_machine_package": "g3-standard-0.5-smartos",
  "source_machine_image": "70e3ae72-96b6-11e6-9056-9737fd4d0764",
  "image_name": "my_new_image",
  "image_version": "1.0.0",
}],
```
