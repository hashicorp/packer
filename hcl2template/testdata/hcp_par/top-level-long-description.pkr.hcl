source "null" "test" {
  communicator = "none"
}

hcp_packer_registry {
  description = <<EOT
This is a super super super super super super super super super super super super super super super super super super
super super super super super super super super super super super super super super super super super super super super
super super super long description
EOT
}

build {
  name = "bucket-slug"
  sources = ["null.test"]
}
