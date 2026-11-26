# Copyright IBM Corp. 2024, 2026
# SPDX-License-Identifier: BUSL-1.1

variable "from_default" {
  type    = any
  default = "hello"
}

variable "from_varfile" {
  type    = any
  default = "hello"
}

variable "from_env" {
  type    = any
  default = "hello"
}

variable "from_argv" {
  type = any
}

variable "untyped" {
}
