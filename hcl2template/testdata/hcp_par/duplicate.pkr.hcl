build {
  name = "bucket-slug"
  hcp_packer_registry {
    description = ""
    bucket_labels = {
      "foo" = "bar"
    }
  }
  hcp_packer_registry {}
}
