source "null" "test" {
  communicator = "none"
}

hcp_packer_registry {
  bucket_name = "ba"
}

build {
  name = "bucket-slug"
  sources = ["null.test"]
}
