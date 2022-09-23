packer {
  required_version = ">= 1.0.0"
}

source "null" "test" {
  communicator = "none"
}

build {
  sources = ["null.test"]
}
