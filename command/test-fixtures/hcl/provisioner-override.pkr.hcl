source "null" "example1" {
  communicator = "none"
}

source "null" "example2" {
  communicator = "none"
}

build {
  sources = ["source.null.example1", "source.null.example2"]
  provisioner "shell-local" {
    inline = ["echo not overridden"]
    override = {
      example1 = {
        inline = ["echo yes overridden"]
      }
    }
  }
}