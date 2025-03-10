source "null" "test" {
  communicator = "none"
}

hcp_packer_registry {
  bucket_name = "ok-Bucket-name-1"
}

build {
  name = "bucket-slug"
  sources = ["null.test"]
}
