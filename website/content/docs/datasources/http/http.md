---
description: |
  The http Data Source retrieves information from an http endpoint to be used
  during Packer builds
page_title: Http - Data Sources
---

<BadgesHeader>
  <PluginBadge type="official" />
  <PluginBadge type="hcp_packer_ready" />
</BadgesHeader>

# Http Data Source

Type: `http`

The `http` Data Source retrieves information from an http endpoint to be used
during packer builds


## Basic Example

Below is a fully functioning example. It stores information about an image
iteration, which can then be parsed and accessed using HCL tools.

```hcl
data "packer-image-iteration" "hardened-source" {
  bucket = "hardened-ubuntu-16-04"
  channel = "production-stable"
}
```

## Configuration Reference

Configuration options are organized below into two categories: required and
optional. Within each category, the available options are alphabetized and
described.

### Required:

@include 'datasource/http/Config-required.mdx'

@include 'datasource/http/Config-not-required.mdx'

## Datasource outputs

The outputs for this datasource are as follows:

@include 'datasource/http/DatasourceOutput.mdx'
