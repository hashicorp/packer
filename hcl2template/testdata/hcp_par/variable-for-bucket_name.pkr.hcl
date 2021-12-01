variable "bucket" {
  type = string
  default = "variable-bucket-slug"
}
build {
  hcp_packer_registry {
    bucket_name   = var.bucket
  }
  sources = [
    "source.virtualbox-iso.ubuntu-1204",
  ]
}

source "virtualbox-iso" "ubuntu-1204" {
}
