source "null" "test" {
  communicator = "none"
}

build {
  name = "bucket-slug"
  hcp_packer_registry {
    bucket_name = "invalid_bucket"
  }

  sources = ["null.test"]
}
