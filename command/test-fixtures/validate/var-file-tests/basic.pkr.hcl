packer {
  required_version = ">= v1.0.0"
}

variable "test" {
 type = string
 default = null
}

source "file" "chocolate" {
  target = "chocolate.txt"
  content = "chocolate"
}

build {
  sources = ["source.file.chocolate"]
}
