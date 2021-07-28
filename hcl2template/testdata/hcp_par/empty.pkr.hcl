build {
  name = "bucket-slug"
  hcp_packer_registry {}

  sources = [
    "source.virtualbox-iso.ubuntu-1204",
  ]
}

source "virtualbox-iso" "ubuntu-1204" {
}

