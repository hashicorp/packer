source "null" "test" {
  communicator = "none"
}

hcp_packer_registry {
  bucket_name = "ok-Bucket-name-1"
}

build {
  name = "bucket-slug-1"
  sources = ["null.test"]
}
build {
  name = "bucket-slug-2"
  sources = ["null.test"]
}
