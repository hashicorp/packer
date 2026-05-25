# Copyright IBM Corp. 2024, 2026
# SPDX-License-Identifier: BUSL-1.1

data "http" "trusted_ca_certificates" {
  method = "GET"
  url    = local.no_dep
}
