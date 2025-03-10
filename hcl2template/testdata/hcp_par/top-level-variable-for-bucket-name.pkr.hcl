variable "bucket" {
  type = string
  default = "variable-bucket-slug"
}

hcp_packer_registry {
  bucket_name   = var.bucket
}

build {
  sources = [
    "source.virtualbox-iso.ubuntu-1204",
  ]
}

source "virtualbox-iso" "ubuntu-1204" {
}
