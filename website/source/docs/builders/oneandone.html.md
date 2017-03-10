---
description: |
    The 1&1 builder is able to create images for 1&1 cloud.
layout: docs
page_title: 1&1 Builder
...

# 1&1 Builder

Type: `oneandone`

The 1&1 Builder is able to create virtual machines for [1&1](https://www.1and1.com/).

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required

-   `source_image_name` (string) - 1&1 Server Appliance name of type `IMAGE`.

-   `token` (string) - 1&1 REST API Token. This can be specified via environment variable `ONEANDONE_TOKEN`

### Optional

-   `data_center_name` - Name of virtual data center. Possible values "ES", "US", "GB", "DE". Default value "US"
 
-   `disk_size` (string) - Amount of disk space for this image in GB. Defaults to "50"

-   `image_name` (string) - Resulting image. If "image_name" is not provided Packer will generate it

-   `retries` (int) - Number of retries Packer will make status requests while waiting for the build to complete. Default value "600".

-   `url` (string) - Endpoint for the 1&1 REST API. Default URL "https://cloudpanel-api.1and1.com/v1"


## Example

Here is a basic example:

```json
{
   "builders":[
      {
         "type":"oneandone",
         "disk_size":"50",
         "image_name":"test5",
         "source_image_name":"ubuntu1604-64min",
         "ssh_username" :"root"
      }
   ]
}
```