source "null" "test" {
  communicator = "none"
}

hcp_packer_registry {
  bucket_name = "invalid_bucket"
}

build {
  name = "bucket-slug"
  sources = ["null.test"]
}
