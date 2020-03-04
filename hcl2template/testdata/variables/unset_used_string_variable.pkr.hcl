
variable "foo" {
  type = string
}

build {
  sources = [
    "source.null.null-builder${var.foo}",
  ]
}
