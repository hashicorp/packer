
variable "vms_to_build" {
}

locals {
  vms_to_build = var.vms_to_build
  dynamic_map = { 
    for vm in local.vms_to_build : 
      vm => lookup(local, vm, "VM NAME NOT FOUND")
  }
}

