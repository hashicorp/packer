
variable "foo" {
  type = string
}


build {
  sources = [
    "source.null.null-builder",
  ]
}

source "null" "null-builder" {
  communicator = "none"
}
