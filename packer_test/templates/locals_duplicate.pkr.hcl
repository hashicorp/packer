local "test" {
    expression = "two"
    sensitive = true
}

locals {
    test = local.test
}

variable "test" {
    type = string
    default = "home"
}
source "null" "example" {}

build {
    sources = ["source.null.example"]
}
