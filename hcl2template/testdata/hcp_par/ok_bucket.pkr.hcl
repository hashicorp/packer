source "null" "test" {
  communicator = "none"
}

build {
  name = "bucket-slug"
  hcp_packer_registry {
    bucket_name = "ok-Bucket-name-1"
  }

  sources = ["null.test"]
}
