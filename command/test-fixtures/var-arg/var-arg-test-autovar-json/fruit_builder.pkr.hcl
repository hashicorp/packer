
variable "fruit" {
  type = string
}

locals {
  fruit = var.fruit
}

source "null" "builder" {
  communicator = "none"
}

build {
  sources = [
    "source.null.builder",
  ]

  provisioner "shell-local" {
    inline = ["echo ${local.fruit} > ${local.fruit}.txt"]
  }
}
