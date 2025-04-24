source "null" "test" {
  communicator = "none"
}
hcp_packer_registry {
  bucket_name = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
}

build {
  name = "bucket-slug"
  sources = ["null.test"]
}
