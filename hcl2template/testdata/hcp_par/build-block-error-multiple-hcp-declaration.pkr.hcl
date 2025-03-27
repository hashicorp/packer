source "null" "test" {
  communicator = "none"
}

build {
  name = "build1"
  hcp_packer_registry {
    bucket_name = "ok-Bucket-name-1"
  }

  sources = ["null.test"]
}

build {
  name = "build2"
  hcp_packer_registry {
    bucket_name = "ok-Bucket-name-1"
  }

  sources = ["null.test"]
}
