variable "region" {
  type =string
}

invalid

source "amazon-ebs" "example" {
  region = var.region
}

build {
  sources = ["source.amazon-ebs.example"]
}

