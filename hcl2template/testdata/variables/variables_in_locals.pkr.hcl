variable "name_prefix" {
  default = "foo"
}

locals {
  default_name_prefix = "${var.project_name}-web"
  name_prefix         = "${var.name_prefix != "" ? var.name_prefix : local.default_name_prefix}"  // works
  foo  = "${local.default_name_prefix}"
  bar  = local.default_name_prefix
}

locals {
  for_var  = "${local.default_name_prefix}"
  bar_var  = local.default_name_prefix
}

variable "project_name" {
  default = "test"
}