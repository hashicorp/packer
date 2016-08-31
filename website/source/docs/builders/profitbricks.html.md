---
description: |
    The ProfitBricks builder is able to create images for ProfitBricks cloud.
layout: docs
page_title: ProfitBricks Builder
...

# ProfitBricks Builder

Type: `profitbricks`

The ProfitBricks Builder is able to create virtual machines for [ProfitBricks](https://www.profitbricks.com).

-> **Note:** This builder is not supported by HashiCorp. Please visit [ProfitBricks DevOps Central](https://devops.profitbricks.com/) for support. You may file issues on [GitHub](https://github.com/profitbricks/packer-builder-profitbricks/issues).

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required

-   `image` (string) - ProfitBricks volume image 

-   `password` (string) - ProfitBrucks password. This can be specified via environment variable `PROFITBRICKS_PASSWORD', if provided. The value definded in the config has precedence over environemnt variable.

-   `username` (string) - ProfitBricks username. This can be specified via environment variable `PROFITBRICKS_USERNAME', if provided. The value definded in the config has precedence over environemnt variable. 


### Optional

-   `cores` (integer) - Amount of CPU cores to use for this build. Defaults to "4".

-   `disk_size` (string) - Amount of disk space for this image. Defaults to "50gb"

-   `disk_type` (string) - Type of disk to use for this image. Defaults to "HDD".

-   `location` (string) - Defaults to "us/las".

-   `ram` (integer) - Amount of RAM to use for this image. Defalts to "2048".

-   `snapshot_name` (string) - If snapshot name is not provided Packer will generate it

-   `snapshot_password` (string) - Password for the snapshot.

-   `timeout` (string) - An approximate limit on how long Packer will continue making status requests while waiting for the build to complete. Default value 120 seconds.

-   `url` (string) - Endpoint for the ProfitBricks REST API. Default URL "https://api.profitbricks.com/rest/v2"


## Example

Here is a basic example:

```json
{
  "builders": [
    {
      "image": "Ubuntu-16.04",
      "type": "profitbricks",
      "disk_size": "5",
      "snapshot_name": "double",
      "ssh_key_path": "/path/to/private/key",
      "snapshot_password": "test1234",
      "timeout": 100
    }
  ]
}
```