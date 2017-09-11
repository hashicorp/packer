  with Oracle Bare Metal Cloud Services (BMCS). The builder takes an
  Oracle-provided base image, runs any provisioning necessary on the base image
  after launching it, and finally snapshots it creating a reusable custom
  image.
---

# Oracle Bare Metal Cloud Services (BMCS) Builder

Type: `oracle-bmcs`

The `oracle-bmcs` Packer builder is able to create new custom images for use
with [Oracle Bare Metal Cloud Services](https://cloud.oracle.com/en_US/bare-metal-compute)
(BMCS). The builder takes an Oracle-provided base image, runs any provisioning
necessary on the base image after launching it, and finally snapshots it
creating a reusable custom image.

It is recommended that you familiarise yourself with the
[Key Concepts and Terminology](https://docs.us-phoenix-1.oraclecloud.com/Content/GSG/Concepts/concepts.htm)
prior to using this builder if you have not done so already.

The builder _does not_ manage images. Once it creates an image, it is up to you
to use it or delete it.

## Authorization

The Oracle BMCS API requires that requests be signed with the RSA public key
associated with your [IAM](https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/overview.htm)
user account. For a comprehensive example of how to configure the required
authentication see the documentation on
[Required Keys and OCIDs](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/apisigningkey.htm)
([Oracle Cloud IDs](https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm)).

## Configuration Reference

There are many configuration options available for the `oracle-bmcs` builder.
In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required

 -  `availability_domain` (string) - The name of the
    [Availability Domain](https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/regions.htm)
    within which a new instance is launched and provisioned.
    The names of the Availability Domains have a prefix that is specific to
    your [tenancy](https://docs.us-phoenix-1.oraclecloud.com/Content/GSG/Concepts/concepts.htm#two).

    To get a list of the Availability Domains, use the
    [ListAvailabilityDomains](https://docs.us-phoenix-1.oraclecloud.com/api/#/en/identity/latest/AvailabilityDomain/ListAvailabilityDomains)
    operation, which is available in the IAM Service API.

 -  `base_image_ocid` (string) - The OCID of the
    [Oracle-provided base image](https://docs.us-phoenix-1.oraclecloud.com/Content/Compute/References/images.htm)
    to use. This is the unique identifier of the image that will be used to
    launch a new instance and provision it.

    To get a list of the accepted image OCIDs, use the
    [ListImages](https://docs.us-phoenix-1.oraclecloud.com/api/#/en/iaas/latest/Image/ListImages)
    operation available in the Core Services API.

 -  `compartment_ocid` (string) - The OCID of the
    [compartment](https://docs.us-phoenix-1.oraclecloud.com/Content/GSG/Tasks/choosingcompartments.htm)

 -  `fingerprint` (string) - Fingerprint for the BMCS API signing key.
    Overrides value provided by the
    [BMCS config file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.

 -  `shape` (string) - The template that determines the number of
    CPUs, amount of memory, and other resources allocated to a newly created
    instance.

    To get a list of the available shapes, use the
    [ListShapes](https://docs.us-phoenix-1.oraclecloud.com/api/#/en/iaas/20160918/Shape/ListShapes)
    operation available in the Core Services API.

 -  `subnet_ocid` (string) - The name of the subnet within which a new instance
    is launched and provisioned.

    To get a list of your subnets, use the
    [ListSubnets](https://docs.us-phoenix-1.oraclecloud.com/api/#/en/iaas/latest/Subnet/ListSubnets)
    operation available in the Core Services API.

    Note: the subnet must be configured to allow access via your chosen
    [communicator](/docs/templates/communicator.html) (communicator defaults to
    [SSH tcp/22](/docs/templates/communicator.html#ssh_port)).


### Optional

 -  `access_cfg_file` (string) - The path to the
    [BMCS config file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm).
    Defaults to `$HOME/.oraclebmc/config`.

 -  `access_cfg_file_account` (string) - The specific account in the
    [BMCS config file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    to use. Defaults to `DEFAULT`.

 -  `image_name` (string) - The name to assign to the resulting custom image.

 -  `key_file` (string) - Full path and filename of the BMCS API signing key.
    Overrides value provided by the
    [BMCS config file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.

 -  `pass_phrase` (string) - Pass phrase used to decrypt the BMCS API signing
    key. Overrides value provided by the
    [BMCS config file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.

 -  `region` (string) - An Oracle Bare Metal Cloud Services region. Overrides
    value provided by the
    [BMCS config file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.

 -  `tenancy_ocid` (string) - The OCID of your tenancy. Overrides value provided
    by the
    [BMCS config file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.

 -  `user_ocid` (string) - The OCID of the user calling the BMCS API. Overrides
    value provided by the [BMCS config file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.


## Basic Example

Here is a basic example. Note that account specific configuration has been
substituted with the letter `a` and OCIDS have been shortened for brevity.

``` {.javascript}
{
  "type": "oracle-bmcs",
  "compartment_ocid": "ocid1.compartment.oc1..aaa",
  "availability_domain": "aaaa:PHX-AD-1",
  "subnet_ocid": "ocid1.subnet.oc1..aaa",
  "base_image_ocid": "ocid1.image.oc1.phx.aaaaaaaa5yu6pw3riqtuhxzov7fdngi4tsteganmao54nq3pyxu3hxcuzmoa",
  "ssh_username": "opc",
  "shape": "VM.Standard1.1",
  "image_name": "ExampleImage"
}
```
