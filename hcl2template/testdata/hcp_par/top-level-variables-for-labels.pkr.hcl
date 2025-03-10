
variable "bucket_labels" {
  type = map(string)
  default = {
    "team": "development",
  }
}

variable "build_labels" {
  type = map(string)
  default = {
    "packageA": "v3.17.5",
    "packageZ": "v0.6",
  }
}

hcp_packer_registry {
  bucket_name   = "bucket-slug"
  bucket_labels = var.bucket_labels
  build_labels  = var.build_labels
}

build {
  sources = [
    "source.virtualbox-iso.ubuntu-1204",
  ]
}

source "virtualbox-iso" "ubuntu-1204" {
}

