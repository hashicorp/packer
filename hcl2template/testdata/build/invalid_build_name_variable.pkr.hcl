variable "name" {
  type    = int
  default = 123
}

build {
  name        = var.name
}

