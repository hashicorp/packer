# Copyright IBM Corp. 2024, 2026
# SPDX-License-Identifier: BUSL-1.1

locals {
  test_local = can(local.test_data) ? local.test_data : []

  test_data = [
    { key = "value" }
  ]
}
