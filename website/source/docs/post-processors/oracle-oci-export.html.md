---
description: |
    The Oracle Cloud infrastructure Image Exporter post-processor exports an image from a Packer oracle-oci builder run and uploads it to oracle cloud object storage. The exported
    images can be easily shared and uploaded to other oracle Cloud Projects.
layout: docs
page_title: 'Oracle cloud Image Exporter - Post-Processors'
sidebar_current: 'docs-post-processors-oralce-oci-export'
---

# Oracle Cloud Infrastructure Image Exporter Post-Processor

Type: `oracle-oci-export`

The Oracle Cloud infrastructure Image Exporter  post-processor exports the resultant image
from a oracle-oci build to a object storage bucket.

The exporter uses the same credentials as oracle-oci builder.



## Configuration

### Required

-   `availability_domain` (string) - The name of the [Availability
    Domain](https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/regions.htm)
    within which a new instance is launched and provisioned. The names of the
    Availability Domains have a prefix that is specific to your
    [tenancy](https://docs.us-phoenix-1.oraclecloud.com/Content/GSG/Concepts/concepts.htm#two).

    To get a list of the Availability Domains, use the
    [ListAvailabilityDomains](https://docs.us-phoenix-1.oraclecloud.com/api/#/en/identity/latest/AvailabilityDomain/ListAvailabilityDomains)
    operation, which is available in the IAM Service API.

-   `compartment_ocid` (string) - The OCID of the
    [compartment](https://docs.us-phoenix-1.oraclecloud.com/Content/GSG/Tasks/choosingcompartments.htm)

-   `fingerprint` (string) - Fingerprint for the OCI API signing key. Overrides
    value provided by the [OCI config
    file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.

### Optional

-   `access_cfg_file` (string) - The path to the [OCI config
    file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm).
    Defaults to `$HOME/.oci/config`.

-   `access_cfg_file_account` (string) - The specific account in the [OCI
    config
    file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    to use. Defaults to `DEFAULT`.

-   `bucket_name` (string) - Bucket name where the image is stored in objectstorage.

-   `image_name` (string) - Name of the image.

-   `key_file` (string) - Full path and filename of the OCI API signing key.
    Overrides value provided by the [OCI config
    file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.

-   `pass_phrase` (string) - Pass phrase used to decrypt the OCI API signing
    key. Overrides value provided by the [OCI config
    file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.

-   `region` (string) - An Oracle Cloud Infrastructure region. Overrides value
    provided by the [OCI config
    file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.

-   `tenancy_ocid` (string) - The OCID of your tenancy. Overrides value
    provided by the [OCI config
    file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.

-   `user_ocid` (string) - The OCID of the user calling the OCI API. Overrides
    value provided by the [OCI config
    file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm)
    if present.

-   `tags` (map of strings) - Add one or more freeform tags to the resulting
    custom image. See [the Oracle
    docs](https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/taggingoverview.htm)
    for more details. Example:

## Basic Example

The following example builds a Oracle cloud infrastructre image and exports the image to a specified bucked in oracle cloud objectstorage.

``` json
{
  "builders": [
    {
        "availability_domain": "aaaa:PHX-AD-1",
        "base_image_ocid": "ocid1.image.oc1.phx.aaaaaaaa5yu6pw3riqtuhxzov7fdngi4tsteganmao54nq3pyxu3hxcuzmoa",
        "compartment_ocid": "ocid1.compartment.oc1..aaa",
        "image_name": "ExampleImage",
        "shape": "VM.Standard1.1",
        "ssh_username": "opc",
        "subnet_ocid": "ocid1.subnet.oc1..aaa",
        "type": "oracle-oci"
    }
  ],
  "post-processors": [
    {
          "type": "oracle-oci-export",
          "availability_domain": "aaaa:PHX-AD-1",
          "compartment_ocid": "ocid1.compartment.oc1..aaa",
          "image_name": "customOracle_V1.2",
          "bucket_name":"ocitest"
   }
  ]
}
```
