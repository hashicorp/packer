source "null" "test" {
  communicator = "none"
}

build {
  name = "bucket-slug"
  hcp_packer_registry {
    description = <<EOT
This is a super super super super super super super super super super super super super super super super super super
super super super super super super super super super super super super super super super super super super super super
super super super long description
    EOT
  }

  sources = ["null.test"]
}
