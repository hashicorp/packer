# Copyright IBM Corp. 2024, 2026
# SPDX-License-Identifier: BUSL-1.1

data "http" "trusted_ca_certificates" {
  method = "GET"
  url    = "http://example.com/ca-bundle.crt"
}

locals {
  test = data.trusted_ca_certificates.url
}
