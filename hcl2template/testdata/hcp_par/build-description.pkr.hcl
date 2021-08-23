build {
  description = <<EOT
Some build description
    EOT
  hcp_packer_registry {
    bucket_name = "bucket-slug"
  }
  sources = [
    "source.virtualbox-iso.ubuntu-1204",
  ]
}

source "virtualbox-iso" "ubuntu-1204" {
}

