variable "region" {
  type =string
}

source "amazon-ebs" "example" {
  region = var.region
}

build {
  sources = ["source.amazon-ebs.example"]
}
