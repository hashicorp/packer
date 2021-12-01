variable "name" {
  type    = string
  default = "example"
}

variable "description" {
  type    = string
  default = "blah blah blah"
}

source "null" "example" {
  communicator = "none"
}

build {
  name        = var.name
  description = var.description

  sources     = ["source.null.example"]

  post-processor "shell-local" {
    inline = ["echo 2 > ${build.name}.2.txt"]
  }
}
