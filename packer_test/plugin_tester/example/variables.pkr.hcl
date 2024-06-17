# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

locals {
  foo = data.scaffolding-my-datasource.mock-data.foo
  bar = data.scaffolding-my-datasource.mock-data.bar
}