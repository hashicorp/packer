source "null" "test" {
  communicator = "none"
}

hcp_packer_registry {
}

build {
  name = "bucket-slug"
  sources = ["null.test"]
}
