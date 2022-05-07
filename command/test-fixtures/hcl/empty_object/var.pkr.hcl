
variable "foo" {
  default = []
}

variable "bar" {
  default = {}
}

source "file" "base" {
}

build {
  source "sources.file.base" {
      target = "${var.bar.baz}.txt"
      content = var.foo[0]
  }
}
