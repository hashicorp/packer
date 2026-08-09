# Copyright IBM Corp. 2024, 2026
# SPDX-License-Identifier: BUSL-1.1

source "null" "test" {
  communicator = "none"
}

build {
  sources = ["null.test"]
}
