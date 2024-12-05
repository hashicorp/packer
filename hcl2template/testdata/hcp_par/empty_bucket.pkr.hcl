source "null" "test" {
  communicator = "none"
}

build {
  name = "bucket-slug"
  hcp_packer_registry {
  }

  sources = ["null.test"]
}
