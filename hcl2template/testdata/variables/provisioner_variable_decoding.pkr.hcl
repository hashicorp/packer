variables {
  max_retries = "1"
  max_retries_int = 1
}

source "null" "null-builder" {
  communicator = "none"
}

build {
  sources = [
    "source.null.null-builder",
  ]

  provisioner "shell" {
    max_retries = var.max_retries_int
  }
  provisioner "shell" {
    max_retries = var.max_retries
  }
}