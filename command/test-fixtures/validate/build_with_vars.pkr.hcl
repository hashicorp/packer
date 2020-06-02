variable "target" {
  type = string
  default = "chocolate.txt"
}

source "file" "chocolate" {
  target = var.target
  content = "chocolate"
}

build {
  sources = ["source.file.chocolate"]
}
