build {
  description = <<EOT
Some build description
    EOT
  hcp_packer_registry {
    bucket_name = "bucket-slug"
    description = <<EOT
Some override description
    EOT
  }
  sources = [
    "source.virtualbox-iso.ubuntu-1204",
  ]
}

source "virtualbox-iso" "ubuntu-1204" {
}

