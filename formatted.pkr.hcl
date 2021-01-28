source "null" "example" {
  communicator = "none"
}

build {
  sources = ["source.null.example"]
}

packer {
  required_plugins {
    comment = {
      source  = "sylviamoss/comment"
      version = "v0.2.14"
    }
  }
}

build {
  sources = ["source.null.example"]

  provisioner "comment-my-provisioner" {
  }
}
