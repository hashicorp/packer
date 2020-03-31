locals {
  name_prefix = "${var.name_prefix}"
  foo         = "${local.name_prefix}"
  bar         = local.name_prefix
}