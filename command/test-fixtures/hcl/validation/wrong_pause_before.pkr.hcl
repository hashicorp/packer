source "null" "example1" {
  communicator = "none"
}

build {
  sources = ["source.null.example1"]

  provisioner "shell-local" {
    pause_before = "5"
    inline       = ["echo Did I wait a bit?"]
  }
}
