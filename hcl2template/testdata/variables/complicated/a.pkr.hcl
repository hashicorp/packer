locals {
  for_var  = "${local.name_prefix}"
  bar_var  = [
    local.for_var,
    local.foo,
    var.name_prefix,
  ]
}