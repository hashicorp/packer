variable "name" {
  type    = string
  default = "build-name"
}

local "description" {
  expression = "This is the description for ${var.name}."
}

build {
  name        = var.name
  description = local.description

  sources = [
    "source.virtualbox-iso.ubuntu-1204"
  ]
}

source "virtualbox-iso" "ubuntu-1204" {
}

