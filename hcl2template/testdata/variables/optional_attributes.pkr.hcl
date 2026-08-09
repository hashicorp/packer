# Copyright IBM Corp. 2024, 2025
# SPDX-License-Identifier: BUSL-1.1

variable "context" {
  type = object({
    with_default    = optional(string, "default_value")
    without_default = optional(number)
    required        = string
  })
  default = {
    required = "set"
  }
}
