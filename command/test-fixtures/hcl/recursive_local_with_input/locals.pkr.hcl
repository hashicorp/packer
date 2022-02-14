variable "vms_to_build" {
  default = {
    "amgroup": "hey"
  }
}

locals {
  vms_to_build = var.vms_to_build
  dynamic_slice = { 
    for vm, val in var.vms_to_build : 
    vm => lookup(local.vms_to_build, vm, "VM NAME NOT FOUND")
  }
}


source "file" "chocolate" {
  content = "hello"
  target = "${local.dynamic_slice.amgroup}.txt"
}

build {
  sources = [
    "sources.file.chocolate",
  ]
}
